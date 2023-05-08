package services

import (
	"reflect"
)

type Service interface {
	Name() string
	OnStart() error
	OnStop() error
	InjectServices() error
}

type CompositeService struct {
	services       []Service
	servicesByName map[string]Service
}

func NewCompositeService(services ...Service) *CompositeService {
	return &CompositeService{
		services:       services,
		servicesByName: make(map[string]Service),
	}
}

func (cs *CompositeService) OnStart() error {
	for _, service := range cs.services {
		err := service.OnStart()
		if err != nil {
			return err
		}
		cs.servicesByName[service.Name()] = service
	}
	return nil
}

func (cs *CompositeService) Test() error {
	for _, service := range cs.services {
		err := service.OnStart()
		if err != nil {
			return err
		}
		cs.servicesByName[service.Name()] = service
	}
	return nil
}

func (cs *CompositeService) OnStop() error {
	for _, service := range cs.services {
		err := service.OnStop()
		if err != nil {
			return err
		}
	}
	return nil
}

func (cs *CompositeService) GetServiceByType(serviceType reflect.Type) Service {
	for _, service := range cs.services {
		if reflect.TypeOf(service) == serviceType {
			return service
		}
	}
	return nil
}
