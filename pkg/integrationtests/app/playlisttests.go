package app

import (
	"github.com/CuriosityMusicStreaming/ComponentsPool/pkg/app/auth"
	"github.com/google/uuid"
)

func playlistTests(playlistServiceAPI PlaylistServiceAPI) {
	createPlaylist(playlistServiceAPI)
	managePlaylist(playlistServiceAPI)
}

func createPlaylist(playlistServiceAPI PlaylistServiceAPI) {
	user := auth.UserDescriptor{UserID: uuid.New()}

	firstPlaylistName := "Gibberish 1000 hours"
	secondPlaylistName := "Chill out"

	{
		firstPlaylistID, err := playlistServiceAPI.CreatePlaylist(firstPlaylistName, user)
		assertNoErr(err)

		playlists, err := playlistServiceAPI.GetUserPlaylists(user)
		assertNoErr(err)

		assertEqual(1, len(playlists.Playlists))

		playlist := playlists.Playlists[0]

		assertEqual(firstPlaylistID, playlist.PlaylistID)
		assertEqual(firstPlaylistName, playlist.Name)
		assertEqual(user.UserID.String(), playlist.OwnerID)

		secondPlaylistID, err := playlistServiceAPI.CreatePlaylist(secondPlaylistName, user)
		assertNoErr(err)

		playlists, err = playlistServiceAPI.GetUserPlaylists(user)
		assertNoErr(err)

		assertEqual(2, len(playlists.Playlists))

		secondPlaylist, err := playlistServiceAPI.GetPlaylist(secondPlaylistID, user)
		assertNoErr(err)

		assertEqual(secondPlaylistName, secondPlaylist.Name)
		assertEqual(user.UserID.String(), secondPlaylist.OwnerID)

		assertNoErr(playlistServiceAPI.DeletePlaylist(firstPlaylistID, user))
		assertNoErr(playlistServiceAPI.DeletePlaylist(secondPlaylistID, user))
	}
}

func managePlaylist(playlistServiceAPI PlaylistServiceAPI) {
	user := auth.UserDescriptor{UserID: uuid.New()}
	anotherUser := auth.UserDescriptor{UserID: uuid.New()}
	playlistName := "Gibberish 1000 hours"
	newPlaylistName := "new title"

	{
		playlistID, err := playlistServiceAPI.CreatePlaylist(playlistName, user)
		assertNoErr(err)

		assertNoErr(playlistServiceAPI.SetPlaylistTitle(playlistID, newPlaylistName, user))

		playlist, err := playlistServiceAPI.GetPlaylist(playlistID, user)
		assertNoErr(err)

		assertEqual(newPlaylistName, playlist.Name)

		assertEqual(playlistServiceAPI.SetPlaylistTitle(playlistID, newPlaylistName, anotherUser), ErrOnlyOwnerCanManagePlaylist)

		playlist, err = playlistServiceAPI.GetPlaylist(playlistID, user)
		assertNoErr(err)

		assertEqual(newPlaylistName, playlist.Name)

		assertEqual(playlistServiceAPI.DeletePlaylist(playlistID, anotherUser), ErrOnlyOwnerCanManagePlaylist)

		assertNoErr(playlistServiceAPI.DeletePlaylist(playlistID, user))
	}
}
