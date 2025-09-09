package models

import "fmt"

type SchoolID int

const (
	HKR SchoolID = iota
	MAU
	ORU
	LTU
	HIG
	SH
	HV
	HB
	MDH
)

func (s SchoolID) String() string {
	names := map[SchoolID]string{
		HKR: "HKR",
		MAU: "MAU",
		ORU: "ORU",
		LTU: "LTU",
		HIG: "HIG",
		SH:  "SH",
		HV:  "HV",
		HB:  "HB",
		MDH: "MDH",
	}
	if name, exists := names[s]; exists {
		return name
	}
	return "Unknown"
}

func (s SchoolID) IsValid() bool {
	return s >= HKR && s <= MDH
}

func ParseSchoolID(s string) (SchoolID, error) {
	schools := map[string]SchoolID{
		"HKR": HKR,
		"MAU": MAU,
		"ORU": ORU,
		"LTU": LTU,
		"HIG": HIG,
		"SH":  SH,
		"HV":  HV,
		"HB":  HB,
		"MDH": MDH,
	}

	if school, exists := schools[s]; exists {
		return school, nil
	}
	return 0, fmt.Errorf("invalid school ID: %s", s)
}

type School struct {
	ID            SchoolID `json:"id"`
	Name          string   `json:"name"`
	URLs          []string `json:"urls"`
	LoginRequired bool     `json:"login_required"`
}

func NewSchool(id SchoolID, name string, urls []string, loginRequired bool) *School {
	return &School{
		ID:            id,
		Name:          name,
		URLs:          urls,
		LoginRequired: loginRequired,
	}
}

func (s *School) String() string {
	return fmt.Sprintf("School{ID: %s, Name: %s, URLs: %d, LoginRequired: %t}",
		s.ID.String(), s.Name, len(s.URLs), s.LoginRequired)
}

func (s *School) HasMultipleURLs() bool {
	return len(s.URLs) > 1
}

func (s *School) GetPrimaryURL() string {
	if len(s.URLs) > 0 {
		return s.URLs[0]
	}
	return ""
}
