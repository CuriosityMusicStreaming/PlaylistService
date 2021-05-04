package domain

type Event interface {
	ID() string
}

type EventHandler interface {
	Handle(event Event) error
}

type EventDispatcher interface {
	Dispatch(event Event) error
}

type EventSource interface {
	Subscribe(handler EventHandler)
}

type EventPublisher interface {
	EventDispatcher
	EventSource
}

func NewEventPublisher() EventPublisher {
	return &eventPublisher{}
}

type eventPublisher struct {
	subscribers []EventHandler
}

func (e *eventPublisher) Dispatch(event Event) error {
	for _, subscriber := range e.subscribers {
		err := subscriber.Handle(event)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *eventPublisher) Subscribe(handler EventHandler) {
	e.subscribers = append(e.subscribers, handler)
}

type PlaylistCreated struct {
	PlaylistID PlaylistID
	OwnerID    PlaylistOwnerID
}

func (p PlaylistCreated) ID() string {
	return "playlist_created"
}

type PlaylistNameChanged struct {
	PlaylistID PlaylistID
	NewName    string
}

func (p PlaylistNameChanged) ID() string {
	return "playlist_name_changed"
}

type PlaylistItemAdded struct {
	PlaylistID     PlaylistID
	PlaylistItemID PlaylistItemID
	ContentID      ContentID
}

func (p PlaylistItemAdded) ID() string {
	return "playlist_item_added"
}

type PlaylistItemRemoved struct {
	PlaylistID     PlaylistID
	PlaylistItemID PlaylistItemID
}

func (p PlaylistItemRemoved) ID() string {
	panic("playlist_item_removed")
}

type PlaylistRemoved struct {
	PlaylistID PlaylistID
	OwnerID    PlaylistOwnerID
}

func (p PlaylistRemoved) ID() string {
	panic("playlist_removed")
}
