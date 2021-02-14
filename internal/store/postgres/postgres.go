package postgres

import (
	"errors"
	"net"

	"github.com/IktaS/go-home/internal/device"
	"github.com/IktaS/go-serv/pkg/serv"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DeviceModel defines a model for the Device struct for use with gorm
type DeviceModel struct {
	DeviceID uuid.UUID `gorm:"column:id;primary_key"`
	Name     string
	Addr     string
	Services []*ServiceModel
	Messages []*MessageModel
}

// ServiceModel defines a model for the Service struct for use with gorm
type ServiceModel struct {
	gorm.Model
	Name     string
	Request  []*TypeModel
	Response *TypeModel
}

// MessageModel defines a model for the Message struct for use with gorm
type MessageModel struct {
	gorm.Model
	Name        string
	Definitions []*MessageDefinitionModel
}

// MessageDefinitionModel defines a model for the MessageDefinition struct for use with gorm
type MessageDefinitionModel struct {
	gorm.Model
	Name       string
	IsOptional bool
	IsRequired bool
	Type       *TypeModel
}

// TypeModel defines a model for the Type struct for use with gorm
type TypeModel struct {
	gorm.Model
	IsScalar  bool
	TypeValue string
}

//Store defines what the Postgre SQL Store needs
type Store struct {
	DSN string
	DB  *gorm.DB
}

func typeArrayToModel(types []*serv.Type) []*TypeModel {
	var typeModels []*TypeModel
	for _, t := range types {
		typeModels = append(typeModels, typeToModel(t))
	}
	return typeModels
}

func typeToModel(t *serv.Type) *TypeModel {
	if t.Reference == "" {
		return &TypeModel{
			IsScalar:  true,
			TypeValue: t.Scalar.String(),
		}
	}
	return &TypeModel{
		IsScalar:  false,
		TypeValue: t.Reference,
	}
}

func messageDefinitionArrayToModel(messageDef []*serv.MessageDefinition) []*MessageDefinitionModel {
	var messageDefModel []*MessageDefinitionModel
	for _, t := range messageDef {
		messageDefModel = append(messageDefModel, messageDefinitionToModel(t))
	}
	return messageDefModel
}

func messageDefinitionToModel(m *serv.MessageDefinition) *MessageDefinitionModel {
	return &MessageDefinitionModel{
		IsOptional: m.Field.Optional,
		IsRequired: m.Field.Required,
		Type:       typeToModel(m.Field.Type),
	}
}

func modelToTypeArray(models []*TypeModel) []*serv.Type {
	var types []*serv.Type
	for _, m := range models {
		types = append(types, modelToType(m))
	}
	return types
}

func modelToType(model *TypeModel) *serv.Type {
	if model.IsScalar {
		return &serv.Type{
			Scalar: serv.StringToScalar[model.TypeValue],
		}
	}
	return &serv.Type{
		Reference: model.TypeValue,
	}
}

func modelToMessageDefinitionArray(models []*MessageDefinitionModel) []*serv.MessageDefinition {
	var mesDef []*serv.MessageDefinition
	for _, m := range models {
		mesDef = append(mesDef, modelToMessageDefinition(m))
	}
	return mesDef
}

func modelToMessageDefinition(model *MessageDefinitionModel) *serv.MessageDefinition {
	return &serv.MessageDefinition{
		Field: &serv.Field{
			Optional: model.IsOptional,
			Required: model.IsRequired,
			Name:     model.Name,
			Type:     modelToType(model.Type),
		},
	}
}
func servicesToModel(services []*serv.Service) []*ServiceModel {
	var models []*ServiceModel

	for _, s := range services {
		models = append(models, &ServiceModel{
			Name:     s.Name,
			Request:  typeArrayToModel(s.Request),
			Response: typeToModel(s.Response),
		})
	}
	return models
}
func modelToServices(models []*ServiceModel) []*serv.Service {
	var services []*serv.Service

	for _, m := range models {
		services = append(services, &serv.Service{
			Name:     m.Name,
			Request:  modelToTypeArray(m.Request),
			Response: modelToType(m.Response),
		})
	}
	return services
}

func messagesToModel(messages []*serv.Message) []*MessageModel {
	var models []*MessageModel

	for _, m := range messages {
		models = append(models, &MessageModel{
			Name:        m.Name,
			Definitions: messageDefinitionArrayToModel(m.Definitions),
		})
	}
	return models
}

func modelToMessages(models []*MessageModel) []*serv.Message {
	var messages []*serv.Message

	for _, m := range models {
		messages = append(messages, &serv.Message{
			Name:        m.Name,
			Definitions: modelToMessageDefinitionArray(m.Definitions),
		})
	}
	return messages
}

func deviceToModel(d *device.Device) *DeviceModel {
	services := servicesToModel(d.Services)
	messages := messagesToModel(d.Messages)

	return &DeviceModel{
		DeviceID: d.ID,
		Name:     d.Name,
		Addr:     d.Addr.String(),
		Services: services,
		Messages: messages,
	}
}

func modelToDevice(model *DeviceModel) *device.Device {
	services := modelToServices(model.Services)
	messages := modelToMessages(model.Messages)

	ip := net.ParseIP(model.Addr)
	addr := &net.IPAddr{IP: ip, Zone: ""}
	return &device.Device{
		ID:       model.DeviceID,
		Name:     model.Name,
		Addr:     addr,
		Services: services,
		Messages: messages,
	}
}

// NewPostgreSQLStore makes a new PostgreSQL Store
func NewPostgreSQLStore(dsn string) (*Store, error) {
	p := &Store{DSN: dsn}
	err := p.Init()
	if err != nil {
		return nil, err
	}
	return p, nil
}

// Init initialize a postgreSQL
func (p *Store) Init() error {
	db, err := gorm.Open(postgres.Open(p.DSN), &gorm.Config{})
	if err != nil {
		return err
	}
	p.DB = db

	p.DB.AutoMigrate(&DeviceModel{})
	return nil
}

// Save saves a device to the postgreSQL store
func (p *Store) Save(d *device.Device) error {
	var dev DeviceModel
	err := p.DB.First(&dev, "DeviceID = ?", d.ID).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		p.DB.Create(deviceToModel(d))
	} else {
		return err
	}
	p.DB.Model(&dev).Updates(deviceToModel(d))
	return nil
}

// Get defines getting a device.Device
func (p *Store) Get(id interface{}) (*device.Device, error) {
	id, ok := id.(uuid.UUID)
	if !ok {
		return nil, errors.New("id needs to be uuid")
	}
	var dev DeviceModel
	err := p.DB.First(&dev, id).Error
	if err != nil {
		return nil, err
	}
	return modelToDevice(&dev), nil
}

// GetAll gets all device
func (p *Store) GetAll() ([]*device.Device, error) {
	var devices []DeviceModel
	err := p.DB.Find(&devices).Error
	if err != nil {
		return nil, err
	}
	var devs []*device.Device
	for _, d := range devices {
		devs = append(devs, modelToDevice(&d))
	}
	return devs, nil
}

// Delete defines getting a device.Device
func (p *Store) Delete(id interface{}) error {
	id, ok := id.(uuid.UUID)
	if !ok {
		return errors.New("id needs to be uuid")
	}
	err := p.DB.Delete(&DeviceModel{}, id).Error
	if err != nil {
		return err
	}
	return nil
}
