package app

import (
	contentserviceapi "playlistservice/api/contentservice"

	"github.com/CuriosityMusicStreaming/ComponentsPool/pkg/app/auth"
	"github.com/google/uuid"
)

func playlistsContentTests(playlistServiceAPI PlaylistServiceAPI, contentServiceAPI ContentServiceAPI, container UserContainer) {
	addToPlaylist(playlistServiceAPI, contentServiceAPI, container)
}

func addToPlaylist(playlistServiceAPI PlaylistServiceAPI, contentServiceAPI ContentServiceAPI, container UserContainer) {
	user := auth.UserDescriptor{UserID: uuid.New()}
	anotherUser := auth.UserDescriptor{UserID: uuid.New()}
	author := auth.UserDescriptor{UserID: uuid.New()}

	container.AddAuthor(author)
	container.AddListener(user)

	var publicContentID string
	var privateContentID string

	resp, err := contentServiceAPI.AddContent(
		"new song",
		contentserviceapi.ContentType_Song,
		contentserviceapi.ContentAvailabilityType_Public,
		author,
	)
	assertNoErr(err)

	publicContentID = resp.ContentID

	resp, err = contentServiceAPI.AddContent(
		"new patreon podcast",
		contentserviceapi.ContentType_Podcast,
		contentserviceapi.ContentAvailabilityType_Private,
		author,
	)
	assertNoErr(err)

	privateContentID = resp.ContentID

	{
		playlistID, err := playlistServiceAPI.CreatePlaylist("collection", user)
		assertNoErr(err)

		playlistItemID, err := playlistServiceAPI.AddToPlaylist(playlistID, publicContentID, user)
		assertNoErr(err)

		playlistResp, err := playlistServiceAPI.GetPlaylist(playlistID, user)
		assertNoErr(err)

		assertEqual(1, len(playlistResp.PlaylistItems))
		assertEqual(playlistItemID, playlistResp.PlaylistItems[0].PlaylistItemID)
		assertEqual(publicContentID, playlistResp.PlaylistItems[0].ContentID)

		_, err = playlistServiceAPI.AddToPlaylist(playlistID, privateContentID, user)
		assertEqual(err, ErrContentNotFound)

		playlistResp, err = playlistServiceAPI.GetPlaylist(playlistID, user)
		assertNoErr(err)

		assertEqual(1, len(playlistResp.PlaylistItems))

		assertEqual(playlistServiceAPI.RemoveFromPlaylist(playlistItemID, anotherUser), ErrOnlyOwnerCanManagePlaylist)

		assertNoErr(playlistServiceAPI.RemoveFromPlaylist(playlistItemID, user))

		assertNoErr(playlistServiceAPI.DeletePlaylist(playlistID, user))
	}

	{
		playlistID, err := playlistServiceAPI.CreatePlaylist("collection", user)
		assertNoErr(err)

		_, err = playlistServiceAPI.AddToPlaylist(playlistID, publicContentID, user)
		assertNoErr(err)

		_, err = playlistServiceAPI.AddToPlaylist(playlistID, publicContentID, user)
		assertNoErr(err)

		assertNoErr(playlistServiceAPI.DeletePlaylist(playlistID, user))
	}
}
