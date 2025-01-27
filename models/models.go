package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var DATABASE *gorm.DB

type Organismo string

const (
	Bacteria    Organismo = "Bacteria"
	Virus       Organismo = "Virus"
	Fungo       Organismo = "Fungo"
	Protozoario Organismo = "Protozoario"
)

func (organismo *Organismo) Scan(value interface{}) error {
	*organismo = Organismo(value.([]byte))
	return nil
}

func (organismo *Organismo) Value() (interface{}, error) {
	return string(*organismo), nil
}

type Zoonose struct {
	gorm.Model
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()"`
	Nome           string    `gorm:"type:text"`
	NomeCientifico string    `gorm:"type:text"`
	Descricao      string    `gorm:"type:text"`
	Organismo      Organismo `gorm:"type:organismo"`
	// FKs
	Agentes      []Agentes      `gorm:"foreignKey:ZoonoseID"`
	Vetores      []Vetores      `gorm:"foreignKey:ZoonoseID"`
	Transmissoes []Transmissoes `gorm:"foreignKey:ZoonoseID"`
	Profilaxias  []Profilaxias  `gorm:"foreignKey:ZoonoseID"`
	Sintomas     []Sintomas     `gorm:"foreignKey:ZoonoseID"`
}

type Agentes struct {
	gorm.Model
	ZoonoseID uuid.UUID `gorm:"index"`
	Agentes   string    `gorm:"type:text"`
}

type Vetores struct {
	gorm.Model
	ZoonoseID uuid.UUID `gorm:"index"`
	Vetores   string    `gorm:"type:text"`
}

type Transmissoes struct {
	gorm.Model
	ZoonoseID    uuid.UUID `gorm:"index"`
	Transmissoes string    `gorm:"type:text"`
}

type Profilaxias struct {
	gorm.Model
	ZoonoseID   uuid.UUID `gorm:"index"`
	Profilaxias string    `gorm:"type:text"`
}

type Sintomas struct {
	gorm.Model
	ZoonoseID uuid.UUID `gorm:"index"`
	Sintomas  string    `gorm:"type:text"`
}

type Aparelhos struct {
	gorm.Model
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()"`
	Nome       string    `gorm:"type:text"`
	ImagePath  string    `gorm:"type:text"`
	VideoPath  string    `gorm:"type:text"`
	ManualPath string    `gorm:"type:text"`
}
