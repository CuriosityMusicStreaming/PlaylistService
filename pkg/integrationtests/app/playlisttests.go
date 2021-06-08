package app

import (
	"github.com/CuriosityMusicStreaming/ComponentsPool/pkg/app/auth"
	"github.com/google/uuid"
)

func playlistTests(playlistServiceApi PlaylistServiceApi, container UserContainer) {
	createPlaylist(playlistServiceApi)
	managePlaylist(playlistServiceApi)
}

func createPlaylist(playlistServiceApi PlaylistServiceApi) {
	user := auth.UserDescriptor{UserID: uuid.New()}

	firstPlaylistName := "Gibberish 1000 hours"
	secondPlaylistName := "Chill out"

	{
		firstPlaylistID, err := playlistServiceApi.CreatePlaylist(firstPlaylistName, user)
		assertNoErr(err)

		playlists, err := playlistServiceApi.GetUserPlaylists(user)
		assertNoErr(err)

		assertEqual(1, len(playlists.Playlists))

		playlist := playlists.Playlists[0]

		assertEqual(firstPlaylistID, playlist.PlaylistID)
		assertEqual(firstPlaylistName, playlist.Name)
		assertEqual(user.UserID.String(), playlist.OwnerID)

		secondPlaylistID, err := playlistServiceApi.CreatePlaylist(secondPlaylistName, user)
		assertNoErr(err)

		playlists, err = playlistServiceApi.GetUserPlaylists(user)
		assertNoErr(err)

		assertEqual(2, len(playlists.Playlists))

		secondPlaylist, err := playlistServiceApi.GetPlaylist(secondPlaylistID, user)
		assertNoErr(err)

		assertEqual(secondPlaylistName, secondPlaylist.Name)
		assertEqual(user.UserID.String(), secondPlaylist.OwnerID)

		assertNoErr(playlistServiceApi.DeletePlaylist(firstPlaylistID, user))
		assertNoErr(playlistServiceApi.DeletePlaylist(secondPlaylistID, user))
	}
}

func managePlaylist(playlistServiceApi PlaylistServiceApi) {
	user := auth.UserDescriptor{UserID: uuid.New()}
	anotherUser := auth.UserDescriptor{UserID: uuid.New()}
	playlistName := "Gibberish 1000 hours"
	newPlaylistName := "new title"

	{
		playlistID, err := playlistServiceApi.CreatePlaylist(playlistName, user)
		assertNoErr(err)

		assertNoErr(playlistServiceApi.SetPlaylistTitle(playlistID, newPlaylistName, user))

		playlist, err := playlistServiceApi.GetPlaylist(playlistID, user)
		assertNoErr(err)

		assertEqual(newPlaylistName, playlist.Name)

		assertEqual(playlistServiceApi.SetPlaylistTitle(playlistID, newPlaylistName, anotherUser), ErrOnlyOwnerCanManagePlaylist)

		playlist, err = playlistServiceApi.GetPlaylist(playlistID, user)
		assertNoErr(err)

		assertEqual(newPlaylistName, playlist.Name)

		assertEqual(playlistServiceApi.DeletePlaylist(playlistID, anotherUser), ErrOnlyOwnerCanManagePlaylist)

		assertNoErr(playlistServiceApi.DeletePlaylist(playlistID, user))
	}
}

func playlistContent(playlistServiceApi PlaylistServiceApi) {

}
