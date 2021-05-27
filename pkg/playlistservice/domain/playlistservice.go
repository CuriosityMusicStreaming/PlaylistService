package domain

import (
	"errors"
)

type PlaylistService interface {
	CreatePlaylist(name string, ownerID PlaylistOwnerID) (PlaylistID, error)
	SetPlaylistName(id PlaylistID, ownerID PlaylistOwnerID, newName string) error
	AddToPlaylist(id PlaylistID, ownerID PlaylistOwnerID, contentID ContentID) (PlaylistItemID, error)
	RemoveFromPlaylist(id PlaylistItemID, ownerID PlaylistOwnerID) error
	RemovePlaylist(id PlaylistID, ownerID PlaylistOwnerID) error
}

func NewPlaylistService(playlistRepo PlaylistRepository, eventDispatcher EventDispatcher) PlaylistService {
	return &playlistService{
		playlistRepo:    playlistRepo,
		eventDispatcher: eventDispatcher,
	}
}

var (
	ErrOnlyOwnerCanManagePlaylist = errors.New("only owner can manage playlist")
)

type playlistService struct {
	playlistRepo    PlaylistRepository
	eventDispatcher EventDispatcher
}

func (service *playlistService) CreatePlaylist(name string, ownerID PlaylistOwnerID) (PlaylistID, error) {
	playlistID := service.playlistRepo.NewID()

	playlist := NewPlaylist(playlistID, name, ownerID)

	err := service.playlistRepo.Store(playlist)
	if err != nil {
		return [16]byte{}, err
	}

	err = service.eventDispatcher.Dispatch(PlaylistCreated{
		PlaylistID: playlist.ID,
		OwnerID:    ownerID,
	})
	if err != nil {
		return [16]byte{}, err
	}

	return playlistID, err
}

func (service *playlistService) SetPlaylistName(id PlaylistID, ownerID PlaylistOwnerID, newName string) error {
	playlist, err := service.playlistRepo.Find(id)
	if err != nil {
		return err
	}

	if playlist.OwnerID != ownerID {
		return ErrOnlyOwnerCanManagePlaylist
	}

	if playlist.Name == newName {
		return nil
	}

	playlist.Name = newName

	err = service.playlistRepo.Store(playlist)
	if err != nil {
		return err
	}

	return service.eventDispatcher.Dispatch(PlaylistNameChanged{PlaylistID: id, NewName: newName})
}

func (service *playlistService) AddToPlaylist(id PlaylistID, ownerID PlaylistOwnerID, contentID ContentID) (PlaylistItemID, error) {
	playlist, err := service.playlistRepo.Find(id)
	if err != nil {
		return [16]byte{}, err
	}

	if playlist.OwnerID != ownerID {
		return [16]byte{}, ErrOnlyOwnerCanManagePlaylist
	}

	playlistItem := PlaylistItem{
		ID:        service.playlistRepo.NewPlaylistItemID(),
		ContentID: contentID,
	}

	playlist.AddItem(playlistItem)

	err = service.playlistRepo.Store(playlist)
	if err != nil {
		return [16]byte{}, err
	}

	err = service.eventDispatcher.Dispatch(PlaylistItemAdded{
		PlaylistID:     playlist.ID,
		PlaylistItemID: playlistItem.ID,
		ContentID:      playlistItem.ContentID,
	})
	if err != nil {
		return [16]byte{}, err
	}

	return playlistItem.ID, nil
}

func (service *playlistService) RemoveFromPlaylist(id PlaylistItemID, ownerID PlaylistOwnerID) error {
	playlist, err := service.playlistRepo.FindByItemID(id)
	if err != nil {
		return err
	}

	if playlist.OwnerID != ownerID {
		return ErrOnlyOwnerCanManagePlaylist
	}

	err = playlist.RemoveItem(id)
	if err != nil {
		return err
	}

	err = service.playlistRepo.Store(playlist)
	if err != nil {
		return err
	}

	return service.eventDispatcher.Dispatch(PlaylistItemRemoved{
		PlaylistID:     playlist.ID,
		PlaylistItemID: id,
	})
}

func (service *playlistService) RemovePlaylist(id PlaylistID, ownerID PlaylistOwnerID) error {
	playlist, err := service.playlistRepo.Find(id)
	if err != nil {
		return err
	}

	if playlist.OwnerID != ownerID {
		return ErrOnlyOwnerCanManagePlaylist
	}

	err = service.playlistRepo.Remove(id)
	if err != nil {
		return err
	}

	return service.eventDispatcher.Dispatch(PlaylistRemoved{
		PlaylistID: id,
		OwnerID:    ownerID,
	})
}
