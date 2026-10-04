package domain

import "testing"

func TestDiagnosticValidation(t *testing.T) {
	for _, value := range []string{"NaN", "Inf", "-Inf", "1e9999", "hello"} {
		if ValidateDiagnosticNumber(value) {
			t.Fatal("invalid numeric result accepted", value)
		}
	}
	for _, value := range []string{"0", "-2.5", "1e2"} {
		if !ValidateDiagnosticNumber(value) {
			t.Fatal("valid numeric result rejected", value)
		}
	}
	input := DiagnosticResultInput{Version: 1, Summary: "Test", Values: []DiagnosticValue{{Position: 1, Value: "a"}, {Position: 1, Value: "b"}}}
	if input.Validate() == nil {
		t.Fatal("duplicate result positions")
	}
	test := DiagnosticTestInput{Kind: "pathology", Name: "Test", ShortName: "T", Category: "Test", Parameters: []DiagnosticParameter{{Name: "value", ValueType: "number"}}}
	if e := test.Validate(); e != nil {
		t.Fatal(e)
	}
	test.Parameters = append(test.Parameters, test.Parameters[0])
	if test.Validate() == nil {
		t.Fatal("duplicate parameter names")
	}
	action := DiagnosticAction{Version: 1, Action: "collect"}
	if action.Validate() == nil {
		t.Fatal("collection without accession reference")
	}
	for _, role := range []string{"admin", "doctor", "patient", "nurse", "receptionist", "pharmacist", "accountant", "case_manager", "lab_technician"} {
		a := Actor{Role: role}
		if a.Can("diagnostics.read") != (role == "admin" || role == "doctor" || role == "patient" || role == "lab_technician") {
			t.Fatal("diagnostic permission", role)
		}
	}
}
