package domain

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPlaylistService_CreatePlaylist(t *testing.T) {
	playlistRepo := newMockPlaylistRepo()
	eventDispatcher := newMockEventDispatcher()

	playlistService := NewPlaylistService(playlistRepo, eventDispatcher)

	{
		newPlaylistName := "magic"
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
		playlistName := "magic"
		playlistOwner := PlaylistOwnerID(uuid.New())

		playlistID, err := playlistService.CreatePlaylist(playlistName, playlistOwner)
		assert.NoError(t, err)

		newPlaylistName := "new-magic"
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
		playlistName := "magic"
		playlistOwner := PlaylistOwnerID(uuid.New())
		anotherPlaylistOwner := PlaylistOwnerID(uuid.New())

		playlistID, err := playlistService.CreatePlaylist(playlistName, playlistOwner)
		assert.NoError(t, err)

		newPlaylistName := "new-magic"
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
		playlistName := "magic"
		playlistOwner := PlaylistOwnerID(uuid.New())
		content := ContentID(uuid.New())

		playlistID, err := playlistService.CreatePlaylist(playlistName, playlistOwner)
		assert.NoError(t, err)

		playlistItemId, err := playlistService.AddToPlaylist(playlistID, playlistOwner, content)
		assert.NoError(t, err)

		playlist, ok := playlistRepo.playlists[playlistID]
		assert.Equal(t, true, ok)

		playlistItem, ok := playlist.Items()[playlistItemId]
		assert.Equal(t, true, ok)

		assert.Equal(t, content, playlistItem.ContentID())

		assert.Equal(t, len(eventDispatcher.events), 2)
		assert.IsType(t, PlaylistItemAdded{}, eventDispatcher.events[1])

		anotherPlaylistOwner := PlaylistOwnerID(uuid.New())
		_, err = playlistService.AddToPlaylist(playlistID, anotherPlaylistOwner, content)
		assert.EqualError(t, err, ErrOnlyOwnerCanManagePlaylist.Error())
		assert.Equal(t, len(eventDispatcher.events), 2)
	}
}

func TestPlaylistService_RemoveFromPlaylist(t *testing.T) {
	playlistRepo := newMockPlaylistRepo()
	eventDispatcher := newMockEventDispatcher()

	playlistService := NewPlaylistService(playlistRepo, eventDispatcher)

	{
		playlistName := "magic"
		playlistOwner := PlaylistOwnerID(uuid.New())
		anotherPlaylistOwner := PlaylistOwnerID(uuid.New())
		content := ContentID(uuid.New())

		playlistID, err := playlistService.CreatePlaylist(playlistName, playlistOwner)
		assert.NoError(t, err)

		playlistItemId, err := playlistService.AddToPlaylist(playlistID, playlistOwner, content)
		assert.NoError(t, err)

		err = playlistService.RemoveFromPlaylist(playlistItemId, anotherPlaylistOwner)
		assert.EqualError(t, err, ErrOnlyOwnerCanManagePlaylist.Error())

		err = playlistService.RemoveFromPlaylist(playlistItemId, playlistOwner)
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
		playlistName := "magic"
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
		map[PlaylistItemID]PlaylistItem{},
	}
}

type mockPlaylistRepository struct {
	playlists     map[PlaylistID]Playlist
	playlistItems map[PlaylistItemID]PlaylistItem
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

func (m *mockPlaylistRepository) FindByItemID(playlistItemId PlaylistItemID) (Playlist, error) {
	for _, playlist := range m.playlists {
		for id := range playlist.Items() {
			if id == playlistItemId {
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
