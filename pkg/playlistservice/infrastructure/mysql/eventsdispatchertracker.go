package mysql

import (
	"database/sql"
	"sync"

	"github.com/CuriosityMusicStreaming/ComponentsPool/pkg/app/storedevent"
	"github.com/CuriosityMusicStreaming/ComponentsPool/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const dispatchTrackerLockName = "dispatch-tracker-lock"

var ErrLockNotAcquired = errors.New("lock for dispatch tracker not acquired")

func NewEventsDispatchTracker(client mysql.TransactionalClient) storedevent.EventsDispatchTracker {
	return &eventsDispatchTracker{client: client}
}

type eventsDispatchTracker struct {
	client      mysql.TransactionalClient
	mutex       sync.Mutex
	transaction mysql.Transaction
	lock        *mysql.Lock
}

func (tracker *eventsDispatchTracker) TrackLastID(transportName string, id storedevent.ID) error {
	const insertQuery = `
		INSERT INTO tracked_stored_event (transport_name, last_stored_event_id, created_at) VALUES (?, ?, now())
		ON DUPLICATE KEY UPDATE last_stored_event_id=VALUES(last_stored_event_id)
	`
	binaryID, err := uuid.UUID(id).MarshalBinary()
	if err != nil {
		return err
	}

	_, err = tracker.client.Exec(insertQuery, transportName, binaryID)
	return err
}

func (tracker *eventsDispatchTracker) LastID(transportName string) (*storedevent.ID, error) {
	const selectQuery = `SELECT last_stored_event_id FROM tracked_stored_event WHERE transport_name = ?`

	var id uuid.UUID
	err := tracker.client.Get(&id, selectQuery, transportName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	ID := storedevent.ID(id)
	return &ID, nil
}

func (tracker *eventsDispatchTracker) Lock() error {
	tracker.mutex.Lock()

	if tracker.transaction == nil {
		var err error
		tracker.transaction, err = tracker.client.BeginTransaction()
		if err != nil {
			return err
		}
	}

	l := mysql.NewLock(tracker.transaction, dispatchTrackerLockName)
	err := l.Lock()
	if err != nil {
		return err
	}

	tracker.lock = &l

	return nil
}

func (tracker *eventsDispatchTracker) Unlock() (err error) {
	defer tracker.mutex.Unlock()

	defer func() {
		if err != nil {
			transactionErr := tracker.transaction.Rollback()
			if transactionErr != nil {
				err = errors.Wrap(err, transactionErr.Error())
			}
		} else {
			err = tracker.transaction.Commit()
		}
		tracker.transaction = nil
	}()

	if tracker.transaction == nil {
		return ErrLockNotAcquired
	}

	err = tracker.lock.Unlock()
	if err != nil {
		return err
	}
	tracker.lock = nil

	return err
}
