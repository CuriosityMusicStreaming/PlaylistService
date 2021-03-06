package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const (
	playlistName = "magic"
)

func TestPlaylistService_CreatePlaylist(t *testing.T) {
	playlistRepo := newMockPlaylistRepo()
	eventDispatcher := newMockEventDispatcher()

	playlistService := NewPlaylistService(playlistRepo, eventDispatcher)

	{
		newPlaylistName := playlistName
		playlistOwner := PlaylistOwnerID(uuid.New())

		playlistID, err := playlistService.CreatePlaylist(newPlaylistName, playlistOwner)
		assert.NoError(t, err)

		assert.Equal(t, len(playlistRepo.playlists), 1, "playlist has been added to repo")

		playlist, err := playlistRepo.Find(playlistID)
		assert.NoError(t, err)

		assert.Equal(t, playlist.ID(), playlistID)
		assert.Equal(t, playlist.Name(), newPlaylistName)
		assert.Equal(t, playlist.OwnerID(), playlistOwner)

		assert.Equal(t, len(eventDispatcher.events), 1)
		assert.IsType(t, PlaylistCreated{}, eventDispatcher.events[0])
	}
}

func TestPlaylistService_SetPlaylistName(t *testing.T) {
	playlistRepo := newMockPlaylistRepo()
	eventDispatcher := newMockEventDispatcher()

	playlistService := NewPlaylistService(playlistRepo, eventDispatcher)

	{
		playlistName := playlistName
		playlistOwner := PlaylistOwnerID(uuid.New())

		playlistID, err := playlistService.CreatePlaylist(playlistName, playlistOwner)
		assert.NoError(t, err)

		newPlaylistName := "new-" + playlistName
		err = playlistService.SetPlaylistName(playlistID, playlistOwner, newPlaylistName)
		assert.NoError(t, err)

		playlist, err := playlistRepo.Find(playlistID)
		assert.NoError(t, err)

		assert.Equal(t, newPlaylistName, playlist.Name())

		assert.Equal(t, len(eventDispatcher.events), 2)
		assert.IsType(t, PlaylistNameChanged{}, eventDispatcher.events[1])

		err = playlistService.SetPlaylistName(playlistID, playlistOwner, newPlaylistName)
		assert.NoError(t, err)

		assert.Equal(t, len(eventDispatcher.events), 2, "when set current name to playlist no event dispatched")
	}

	{
		playlistName := playlistName
		playlistOwner := PlaylistOwnerID(uuid.New())
		anotherPlaylistOwner := PlaylistOwnerID(uuid.New())

		playlistID, err := playlistService.CreatePlaylist(playlistName, playlistOwner)
		assert.NoError(t, err)

		newPlaylistName := "new-" + playlistName
		err = playlistService.SetPlaylistName(playlistID, anotherPlaylistOwner, newPlaylistName)
		assert.EqualError(t, err, ErrOnlyOwnerCanManagePlaylist.Error())

		playlist, err := playlistRepo.Find(playlistID)
		assert.NoError(t, err)

		assert.Equal(t, playlist.Name(), playlistName, "playlist name didnt change")
	}
}

func TestPlaylistService_AddToPlaylist(t *testing.T) {
	playlistRepo := newMockPlaylistRepo()
	eventDispatcher := newMockEventDispatcher()

	playlistService := NewPlaylistService(playlistRepo, eventDispatcher)

	{
		playlistName := playlistName
		playlistOwner := PlaylistOwnerID(uuid.New())
		content := ContentID(uuid.New())

		playlistID, err := playlistService.CreatePlaylist(playlistName, playlistOwner)
		assert.NoError(t, err)

		playlistItemID, err := playlistService.AddToPlaylist(playlistID, playlistOwner, content)
		assert.NoError(t, err)

		playlist, ok := playlistRepo.playlists[playlistID]
		assert.Equal(t, true, ok)

		playlistItem, ok := playlist.Items()[playlistItemID]
		assert.Equal(t, true, ok)

		assert.Equal(t, content, playlistItem.ContentID())

		assert.Equal(t, len(eventDispatcher.events), 2)
		assert.IsType(t, PlaylistItemAdded{}, eventDispatcher.events[1])

		anotherPlaylistOwner := PlaylistOwnerID(uuid.New())
		_, err = playlistService.AddToPlaylist(playlistID, anotherPlaylistOwner, content)
		assert.EqualError(t, err, ErrOnlyOwnerCanManagePlaylist.Error())
		assert.Equal(t, len(eventDispatcher.events), 2)
	}

	{
		playlistRepo = newMockPlaylistRepo()
		playlistService = NewPlaylistService(playlistRepo, eventDispatcher)

		playlistName := playlistName
		playlistOwner := PlaylistOwnerID(uuid.New())
		content1 := ContentID(uuid.New())
		content2 := ContentID(uuid.New())
		content3 := ContentID(uuid.New())

		playlistID, err := playlistService.CreatePlaylist(playlistName, playlistOwner)
		assert.NoError(t, err)

		playlistItemID1, err := playlistService.AddToPlaylist(playlistID, playlistOwner, content1)
		assert.NoError(t, err)

		playlistItemID2, err := playlistService.AddToPlaylist(playlistID, playlistOwner, content2)
		assert.NoError(t, err)

		playlistItemID3, err := playlistService.AddToPlaylist(playlistID, playlistOwner, content3)
		assert.NoError(t, err)

		playlist, ok := playlistRepo.playlists[playlistID]
		assert.Equal(t, true, ok)

		playlistItem1, ok := playlist.Items()[playlistItemID1]
		assert.Equal(t, true, ok)

		playlistItem2, ok := playlist.Items()[playlistItemID2]
		assert.Equal(t, true, ok)

		playlistItem3, ok := playlist.Items()[playlistItemID3]
		assert.Equal(t, true, ok)

		assert.Equal(t, content1, playlistItem1.contentID)
		assert.Equal(t, content2, playlistItem2.contentID)
		assert.Equal(t, content3, playlistItem3.contentID)
	}
}

func TestPlaylistService_RemoveFromPlaylist(t *testing.T) {
	playlistRepo := newMockPlaylistRepo()
	eventDispatcher := newMockEventDispatcher()

	playlistService := NewPlaylistService(playlistRepo, eventDispatcher)

	{
		playlistName := playlistName
		playlistOwner := PlaylistOwnerID(uuid.New())
		anotherPlaylistOwner := PlaylistOwnerID(uuid.New())
		content := ContentID(uuid.New())

		playlistID, err := playlistService.CreatePlaylist(playlistName, playlistOwner)
		assert.NoError(t, err)

		playlistItemID, err := playlistService.AddToPlaylist(playlistID, playlistOwner, content)
		assert.NoError(t, err)

		err = playlistService.RemoveFromPlaylist(playlistItemID, anotherPlaylistOwner)
		assert.EqualError(t, err, ErrOnlyOwnerCanManagePlaylist.Error())

		err = playlistService.RemoveFromPlaylist(playlistItemID, playlistOwner)
		assert.NoError(t, err)

		assert.Equal(t, len(eventDispatcher.events), 3)
		assert.IsType(t, PlaylistItemRemoved{}, eventDispatcher.events[2])
	}
}

func TestPlaylistService_RemovePlaylist(t *testing.T) {
	playlistRepo := newMockPlaylistRepo()
	eventDispatcher := newMockEventDispatcher()

	playlistService := NewPlaylistService(playlistRepo, eventDispatcher)

	{
		playlistName := playlistName
		playlistOwner := PlaylistOwnerID(uuid.New())
		anotherPlaylistOwner := PlaylistOwnerID(uuid.New())

		playlistID, err := playlistService.CreatePlaylist(playlistName, playlistOwner)
		assert.NoError(t, err)

		err = playlistService.RemovePlaylist(playlistID, anotherPlaylistOwner)
		assert.EqualError(t, err, ErrOnlyOwnerCanManagePlaylist.Error())

		assert.Equal(t, len(eventDispatcher.events), 1)

		err = playlistService.RemovePlaylist(playlistID, playlistOwner)
		assert.NoError(t, err)

		assert.Equal(t, len(eventDispatcher.events), 2)
		assert.IsType(t, PlaylistRemoved{}, eventDispatcher.events[1])
	}
}

func newMockPlaylistRepo() *mockPlaylistRepository {
	return &mockPlaylistRepository{
		map[PlaylistID]Playlist{},
	}
}

type mockPlaylistRepository struct {
	playlists map[PlaylistID]Playlist
}

func (m *mockPlaylistRepository) NewID() PlaylistID {
	return PlaylistID(uuid.New())
}

func (m *mockPlaylistRepository) NewPlaylistItemID() PlaylistItemID {
	return PlaylistItemID(uuid.New())
}

func (m *mockPlaylistRepository) Find(id PlaylistID) (Playlist, error) {
	playlist, ok := m.playlists[id]
	if !ok {
		return Playlist{}, ErrPlaylistNotFound
	}

	return playlist, nil
}

func (m *mockPlaylistRepository) FindByItemID(playlistItemID PlaylistItemID) (Playlist, error) {
	for _, playlist := range m.playlists {
		for id := range playlist.Items() {
			if id == playlistItemID {
				return playlist, nil
			}
		}
	}
	return Playlist{}, ErrPlaylistNotFound
}

func (m *mockPlaylistRepository) Store(playlist Playlist) error {
	m.playlists[playlist.ID()] = playlist

	return nil
}

func (m *mockPlaylistRepository) Remove(id PlaylistID) error {
	delete(m.playlists, id)

	return nil
}

func newMockEventDispatcher() *mockEventDispatcher {
	return &mockEventDispatcher{}
}

type mockEventDispatcher struct {
	events []Event
}

func (eventDispatcher *mockEventDispatcher) Dispatch(event Event) error {
	eventDispatcher.events = append(eventDispatcher.events, event)

	return nil
}
