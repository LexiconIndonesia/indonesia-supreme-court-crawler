package models

type UrlFrontierMetadata struct {
	Title            string   `json:"title"`
	Defendant        string   `json:"defendant"`
	DecisionDate     string   `json:"decision_date"`
	RegistrationDate string   `json:"registration_date"`
	UploadDate       string   `json:"upload_date"`
	Breadcrumbs      []string `json:"breadcrumbs"`
}
