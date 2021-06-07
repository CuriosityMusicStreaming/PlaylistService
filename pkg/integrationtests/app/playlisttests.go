package app

import (
	"github.com/CuriosityMusicStreaming/ComponentsPool/pkg/app/auth"
	"github.com/google/uuid"
)

func playlistTests(serviceApiFacade *playlistServiceApiFacade, container UserContainer) {
	createPlaylist(serviceApiFacade)
}

func createPlaylist(serviceApiFacade *playlistServiceApiFacade) {
	user := auth.UserDescriptor{UserID: uuid.New()}

	firstPlaylistName := "Gibberish 1000 hours"
	secondPlaylistName := "Chill out"

	{
		firstPlaylistID, err := serviceApiFacade.CreatePlaylist(firstPlaylistName, user)
		assertNoErr(err)

		playlists, err := serviceApiFacade.GetUserPlaylists(user)
		assertNoErr(err)

		assertEqual(1, len(playlists.Playlists))

		playlist := playlists.Playlists[0]

		assertEqual(firstPlaylistID, playlist.PlaylistID)
		assertEqual(firstPlaylistName, playlist.Name)
		assertEqual(user.UserID.String(), playlist.OwnerID)

		secondPlaylistID, err := serviceApiFacade.CreatePlaylist(secondPlaylistName, user)
		assertNoErr(err)

		playlists, err = serviceApiFacade.GetUserPlaylists(user)
		assertNoErr(err)

		assertEqual(2, len(playlists.Playlists))

		secondPlaylist, err := serviceApiFacade.GetPlaylist(secondPlaylistID, user)
		assertNoErr(err)

		assertEqual(secondPlaylistName, secondPlaylist.Name)
		assertEqual(user.UserID.String(), secondPlaylist.OwnerID)

		assertNoErr(serviceApiFacade.DeletePlaylist(firstPlaylistID, user))
		assertNoErr(serviceApiFacade.DeletePlaylist(secondPlaylistID, user))
	}
}

func managePlaylist(serviceApiFacade *playlistServiceApiFacade) {
	user := auth.UserDescriptor{UserID: uuid.New()}
	//anotherUser := auth.UserDescriptor{UserID: uuid.New()}
	playlistName := "Gibberish 1000 hours"
	newPlaylistName := "new title"

	{
		playlistID, err := serviceApiFacade.CreatePlaylist(playlistName, user)
		assertNoErr(err)

		assertNoErr(serviceApiFacade.SetPlaylistTitle(playlistID, newPlaylistName, user))

		playlist, err := serviceApiFacade.GetPlaylist(playlistID, user)
		assertNoErr(err)

		assertEqual(newPlaylistName, playlist.Name)

		// error happend cause anotherUser cannot manage playlist
		//assertNoErr(serviceApiFacade.SetPlaylistTitle(playlistID, newPlaylistName, anotherUser))

		playlist, err = serviceApiFacade.GetPlaylist(playlistID, user)
		assertNoErr(err)

		assertEqual(newPlaylistName, playlist.Name)
	}
}
