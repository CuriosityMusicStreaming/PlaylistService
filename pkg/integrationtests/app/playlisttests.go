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
