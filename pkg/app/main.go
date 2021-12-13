package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/gocarina/gocsv"
)

type StudentGrade struct {
	//csv headers are matricula,nome,materia,nota,trimestre
	Matricula string  `csv:"matricula"`
	Nome      string  `csv:"nome"`
	Materia   string  `csv:"materia"`
	Nota      float32 `csv:"nota"`
	Trimestre string  `csv:"trimestre"`
}

type ApprovedStudent struct {
	Matricula string `csv:"matricula"`
	Nome      string `csv:"nome"`
}

type DeniedStudent struct {
	Matricula string `csv:"matricula"`
	Nome      string `csv:"nome"`
	Motivo    string `csv:"motivo"`
}

type ReadStudent struct {
	Matricula string
	Nome      string
	Materia   string
	NotaFinal float32
}

func processCsvFiles(fileNames []string) {

	studentgrades := []*StudentGrade{}
	//Read all files from their names
	for _, fileName := range fileNames {
		f, err := os.Open("../in/" + fileName)

		if err != nil {
			log.Fatal("Unable to read input file "+fileName, err)
		}
		defer f.Close()

		fileGrades := []*StudentGrade{}

		if err := gocsv.UnmarshalFile(f, &fileGrades); err != nil { // Load studentgrades from csv file
			panic(err)
		}
		studentgrades = append(studentgrades, fileGrades...)
	}

	//Initialize processed students
	readStudents := []*ReadStudent{}

	// populate readStudents for comparison
	// add the last occurrence always
	for i := len(studentgrades) - 1; i >= 0; i-- {
		if !contains(readStudents, studentgrades[i]) {
			// "append" to the front of the array
			readStudent := ReadStudent{
				Matricula: studentgrades[i].Matricula,
				Nome:      studentgrades[i].Nome,
				Materia:   studentgrades[i].Materia,
				NotaFinal: 0.0}

			readStudents = append(readStudents, &readStudent)
		}
	}

	//Iterate every studentGrade struct
	//Sum all grades
	for _, student := range studentgrades {
		for _, readStudent := range readStudents {
			if student.Matricula == readStudent.Matricula {
				if student.Materia == readStudent.Materia {
					readStudent.NotaFinal += student.Nota
				}
			}
		}
	}

	var deniedStudents []*DeniedStudent
	var approvedStudents []*ApprovedStudent

	//Calculates the average in 4 trimesters (not parameterized) and enrich denied slice
	for _, readStudent := range readStudents {
		readStudent.NotaFinal = readStudent.NotaFinal / 4
		if readStudent.NotaFinal < 5 {
			reason := fmt.Sprint("Aluno reprovado por nota ", float32(readStudent.NotaFinal), " na matéria ", readStudent.Materia, ". ")
			if !isDenied(deniedStudents, readStudent) {
				deniedStudent := DeniedStudent{
					Matricula: readStudent.Matricula,
					Nome:      readStudent.Nome,
					Motivo:    reason}
				deniedStudents = append(deniedStudents, &deniedStudent)
			} else {
				appendReason(deniedStudents, readStudent, reason)
			}
		}
	}

	//Enrich approved slice
	for _, readStudent := range readStudents {
		if !isDenied(deniedStudents, readStudent) {
			approvedStudent := ApprovedStudent{
				Matricula: readStudent.Matricula,
				Nome:      readStudent.Nome,
			}
			approvedStudents = append(approvedStudents, &approvedStudent)
		}
	}

	//Creates the CSV writer
	gocsv.SetCSVWriter(func(out io.Writer) *gocsv.SafeCSVWriter {
		writer := csv.NewWriter(out)
		return gocsv.NewSafeCSVWriter(writer)
	})

	//writes csv file for approved
	approvedFile, err := os.Create("../out/approved.csv")
	if err != nil {
		log.Fatal(err)
	}
	gocsv.MarshalFile(&approvedStudents, approvedFile)

	//writes csv file for denied students
	deniedFile, err := os.Create("../out/denied.csv")
	if err != nil {
		log.Fatal(err)
	}
	gocsv.MarshalFile(&deniedStudents, deniedFile)
}

func appendReason(list []*DeniedStudent, x *ReadStudent, reason string) {
	for i := range list {
		if x.Matricula == list[i].Matricula {
			list[i].Motivo = list[i].Motivo + reason
		}
	}
}

func isDenied(list []*DeniedStudent, x *ReadStudent) bool {
	for i := range list {
		if x.Matricula == list[i].Matricula {
			return true
		}
	}
	return false
}

func contains(list []*ReadStudent, x *StudentGrade) bool {
	for i := range list {
		if x.Matricula == list[i].Matricula {
			if x.Materia == list[i].Materia {
				return true
			}
		}
	}
	return false
}

func getFiles() []string {
	files, err := ioutil.ReadDir("../in")

	if err != nil {
		log.Fatal(err)
	}

	var fileArray []string

	for _, f := range files {
		fileArray = append(fileArray, f.Name())
	}

	return fileArray
}

func main() {
	processCsvFiles(getFiles())
}
