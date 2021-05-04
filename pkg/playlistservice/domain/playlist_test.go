package domain

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPlaylistService_CreatePlaylist(t *testing.T) {
	playlistRepo := newMockPlaylistRepo()
	playlistItemRepo := newMockPlaylistItemRepo()
	eventDispatcher := newMockEventDispatcher()

	playlistService := NewPlaylistService(
		playlistRepo,
		playlistItemRepo,
		eventDispatcher,
	)

	{
		newPlaylistName := "magic"
		playlistOwner := PlaylistOwnerID(uuid.New())

		playlistID, err := playlistService.CreatePlaylist(newPlaylistName, playlistOwner)
		assert.NoError(t, err)

		assert.Equal(t, len(playlistRepo.playlists), 1, "playlist has been added to repo")

		playlist, err := playlistRepo.Find(playlistID)
		assert.NoError(t, err)

		assert.Equal(t, playlist.ID, playlistID)
		assert.Equal(t, playlist.Name, newPlaylistName)
		assert.Equal(t, playlist.OwnerID, playlistOwner)

		assert.Equal(t, len(eventDispatcher.events), 1)
		assert.IsType(t, PlaylistCreated{}, eventDispatcher.events[0])
	}
}

func TestPlaylistService_SetPlaylistName(t *testing.T) {
	playlistRepo := newMockPlaylistRepo()
	playlistItemRepo := newMockPlaylistItemRepo()
	eventDispatcher := newMockEventDispatcher()

	playlistService := NewPlaylistService(
		playlistRepo,
		playlistItemRepo,
		eventDispatcher,
	)

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

		assert.Equal(t, newPlaylistName, playlist.Name)

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

		assert.Equal(t, playlist.Name, playlistName, "playlist name didnt change")
	}
}

func TestPlaylistService_AddToPlaylist(t *testing.T) {
	playlistRepo := newMockPlaylistRepo()
	playlistItemRepo := newMockPlaylistItemRepo()
	eventDispatcher := newMockEventDispatcher()

	playlistService := NewPlaylistService(
		playlistRepo,
		playlistItemRepo,
		eventDispatcher,
	)

	{
		playlistName := "magic"
		playlistOwner := PlaylistOwnerID(uuid.New())
		content := ContentID(uuid.New())

		playlistID, err := playlistService.CreatePlaylist(playlistName, playlistOwner)
		assert.NoError(t, err)

		playlistItemId, err := playlistService.AddToPlaylist(playlistID, playlistOwner, content)
		assert.NoError(t, err)

		playlistItem, err := playlistItemRepo.Find(playlistItemId)
		assert.NoError(t, err)

		assert.Equal(t, content, playlistItem.ContentID)
		assert.Equal(t, playlistID, playlistItem.PlaylistID)

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
	playlistItemRepo := newMockPlaylistItemRepo()
	eventDispatcher := newMockEventDispatcher()

	playlistService := NewPlaylistService(
		playlistRepo,
		playlistItemRepo,
		eventDispatcher,
	)

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
	playlistItemRepo := newMockPlaylistItemRepo()
	eventDispatcher := newMockEventDispatcher()

	playlistService := NewPlaylistService(
		playlistRepo,
		playlistItemRepo,
		eventDispatcher,
	)

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
	return &mockPlaylistRepository{map[PlaylistID]Playlist{}}
}

type mockPlaylistRepository struct {
	playlists map[PlaylistID]Playlist
}

func (m *mockPlaylistRepository) NewID() PlaylistID {
	return PlaylistID(uuid.New())
}

func (m *mockPlaylistRepository) Find(id PlaylistID) (Playlist, error) {
	playlist, ok := m.playlists[id]
	if !ok {
		return Playlist{}, ErrPlaylistNotFound
	}

	return playlist, nil
}

func (m *mockPlaylistRepository) Store(playlist Playlist) error {
	m.playlists[playlist.ID] = playlist

	return nil
}

func (m *mockPlaylistRepository) Remove(id PlaylistID) error {
	delete(m.playlists, id)

	return nil
}

func newMockPlaylistItemRepo() *mockPlaylistItemRepo {
	return &mockPlaylistItemRepo{map[PlaylistItemID]PlaylistItem{}}
}

type mockPlaylistItemRepo struct {
	playlistsItems map[PlaylistItemID]PlaylistItem
}

func (m *mockPlaylistItemRepo) NewID() PlaylistItemID {
	return PlaylistItemID(uuid.New())
}

func (m *mockPlaylistItemRepo) Find(id PlaylistItemID) (PlaylistItem, error) {
	playlistItem, ok := m.playlistsItems[id]
	if !ok {
		return PlaylistItem{}, ErrPlaylistItemNotFound
	}
	return playlistItem, nil
}

func (m *mockPlaylistItemRepo) Store(playlistItem PlaylistItem) error {
	m.playlistsItems[playlistItem.ID] = playlistItem
	return nil
}

func (m *mockPlaylistItemRepo) Remove(id PlaylistItemID) error {
	delete(m.playlistsItems, id)
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
