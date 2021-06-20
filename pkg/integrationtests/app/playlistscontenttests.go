package app

import (
	contentserviceapi "playlistservice/api/contentservice"

	"github.com/CuriosityMusicStreaming/ComponentsPool/pkg/app/auth"
	"github.com/google/uuid"
)

func playlistsContentTests(playlistServiceApi PlaylistServiceApi, contentServiceApi ContentServiceApi, container UserContainer) {
	addToPlaylist(playlistServiceApi, contentServiceApi, container)
}

func addToPlaylist(playlistServiceApi PlaylistServiceApi, contentServiceApi ContentServiceApi, container UserContainer) {
	user := auth.UserDescriptor{UserID: uuid.New()}
	anotherUser := auth.UserDescriptor{UserID: uuid.New()}
	author := auth.UserDescriptor{UserID: uuid.New()}

	container.AddAuthor(author)
	container.AddListener(user)

	var publicContentID string
	var privateContentID string

	resp, err := contentServiceApi.AddContent(
		"new song",
		contentserviceapi.ContentType_Song,
		contentserviceapi.ContentAvailabilityType_Public,
		author,
	)
	assertNoErr(err)

	publicContentID = resp.ContentID

	resp, err = contentServiceApi.AddContent(
		"new patreon podcast",
		contentserviceapi.ContentType_Podcast,
		contentserviceapi.ContentAvailabilityType_Private,
		author,
	)
	assertNoErr(err)

	privateContentID = resp.ContentID

	{
		playlistID, err := playlistServiceApi.CreatePlaylist("collection", user)
		assertNoErr(err)

		playlistItemID, err := playlistServiceApi.AddToPlaylist(playlistID, publicContentID, user)
		assertNoErr(err)

		playlistResp, err := playlistServiceApi.GetPlaylist(playlistID, user)
		assertNoErr(err)

		assertEqual(1, len(playlistResp.PlaylistItems))
		assertEqual(playlistItemID, playlistResp.PlaylistItems[0].PlaylistItemID)
		assertEqual(publicContentID, playlistResp.PlaylistItems[0].ContentID)

		_, err = playlistServiceApi.AddToPlaylist(playlistID, privateContentID, user)
		assertEqual(err, ErrContentNotFound)

		playlistResp, err = playlistServiceApi.GetPlaylist(playlistID, user)
		assertNoErr(err)

		assertEqual(1, len(playlistResp.PlaylistItems))

		assertEqual(playlistServiceApi.RemoveFromPlaylist(playlistItemID, anotherUser), ErrOnlyOwnerCanManagePlaylist)

		assertNoErr(playlistServiceApi.RemoveFromPlaylist(playlistItemID, user))

		assertNoErr(playlistServiceApi.DeletePlaylist(playlistID, user))
	}

	{
		playlistID, err := playlistServiceApi.CreatePlaylist("collection", user)
		assertNoErr(err)

		_, err = playlistServiceApi.AddToPlaylist(playlistID, publicContentID, user)
		assertNoErr(err)

		_, err = playlistServiceApi.AddToPlaylist(playlistID, publicContentID, user)
		assertNoErr(err)

		assertNoErr(playlistServiceApi.DeletePlaylist(playlistID, user))
	}

}
