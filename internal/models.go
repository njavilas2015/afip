package internal

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var VATConditions = []string{
	"IVA Responsable Inscripto",
	"IVA Responsable No Inscripto",
	"IVA Exento",
	"No Responsable IVA",
	"Responsable Monotributo",
}

var ClientVATConditions = []string{
	"IVA Responsable Inscripto",
	"IVA Responsable No Inscripto",
	"IVA Sujeto Exento",
	"Consumidor Final",
	"Responsable Monotributo",
	"Proveedor del Exterior",
	"Cliente del Exterior",
	"IVA Liberado - Ley Nº 19.640",
	"IVA Responsable Inscripto - Agente de Percepción",
	"Monotributista Social",
	"IVA no alcanzado",
}

type GenericAfipType struct {
	ID          uint      `gorm:"primaryKey"`
	Code        string    `gorm:"size:3;not null;unique"`
	Description string    `gorm:"size:250;not null"`
	ValidFrom   time.Time `gorm:"type:date"`
	ValidTo     time.Time `gorm:"type:date"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (GenericAfipType) TableName() string {
	return "generic_afip_types"
}

func (t GenericAfipType) String() string {
	return fmt.Sprintf("Code: %s, Description: %s", t.Code, t.Description)
}

func (t GenericAfipType) NaturalKey() string {
	return t.Code
}

func LoadMetadata(db *gorm.DB) {

	types := []GenericAfipType{
		{Code: "001", Description: "Tipo Genérico 1", ValidFrom: time.Now().AddDate(-1, 0, 0)},
		{Code: "002", Description: "Tipo Genérico 2", ValidFrom: time.Now().AddDate(-2, 0, 0)},
	}

	for _, t := range types {
		db.FirstOrCreate(&t, GenericAfipType{Code: t.Code})
		log.Printf("Cargado: %v\n", t)
	}
}

func CheckResponse(response map[string]interface{}) error {
	if errors, exists := response["Errors"]; exists && len(errors.([]string)) > 0 {
		return fmt.Errorf("AFIP Exception: %v", errors)
	}
	return nil
}

func InitDb() {

	db, err := gorm.Open(sqlite.Open("afip.db"), &gorm.Config{})

	if err != nil {
		log.Fatalf("Error al conectar a la base de datos: %v", err)
	}

	err = db.AutoMigrate(&GenericAfipType{})

	if err != nil {
		log.Fatalf("Error al migrar modelos: %v", err)
	}

	LoadMetadata(db)

	var afipType GenericAfipType

	if err := db.First(&afipType).Error; err == nil {
		log.Printf("Primer registro: %v\n", afipType)
	} else {
		log.Printf("No se encontró ningún registro: %v\n", err)
	}
}
