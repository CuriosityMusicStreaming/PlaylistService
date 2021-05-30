package mysql

import (
	"playlistservice/pkg/playlistservice/app/service"
)

type UnitOfWorkCompleteNotifier func()

func NewNotifyingUnitOfWorkFactory(factory service.UnitOfWorkFactory, completeNotifier UnitOfWorkCompleteNotifier) *notifyingUnitOfWorkFactoryDecorator {
	return &notifyingUnitOfWorkFactoryDecorator{
		factory:          factory,
		completeNotifier: completeNotifier,
	}
}

type notifyingUnitOfWorkFactoryDecorator struct {
	factory          service.UnitOfWorkFactory
	completeNotifier UnitOfWorkCompleteNotifier
}

func (decorator *notifyingUnitOfWorkFactoryDecorator) NewUnitOfWork(lockName string) (service.UnitOfWork, error) {
	unitOfWork, err := decorator.factory.NewUnitOfWork(lockName)
	if err != nil {
		return nil, err
	}

	if decorator.completeNotifier != nil {
		return &notifyingUnitOfWorkDecorator{
			UnitOfWork:       unitOfWork,
			completeNotifier: decorator.completeNotifier,
		}, nil
	}

	return unitOfWork, nil
}

type notifyingUnitOfWorkDecorator struct {
	service.UnitOfWork
	completeNotifier UnitOfWorkCompleteNotifier
}

func (decorator *notifyingUnitOfWorkDecorator) Complete(err error) error {
	err = decorator.UnitOfWork.Complete(err)
	if err == nil {
		decorator.completeNotifier()
	}
	return err
}
