package dataminr

// AlertResponse represents the response from the Dataminr alerts API
type AlertResponse struct {
	Alerts   []Alert `json:"alerts"`
	NextPage string  `json:"nextPage"`
}

// Alert represents a single Dataminr alert
type Alert struct {
	AlertID                string                  `json:"alertId"`
	AlertTimestamp         string                  `json:"alertTimestamp"`
	AlertType              *AlertType              `json:"alertType,omitempty"`
	Headline               string                  `json:"headline"`
	SubHeadline            *SubHeadline            `json:"subHeadline,omitempty"`
	PublicPost             *PublicPost             `json:"publicPost,omitempty"`
	EstimatedEventLocation *EstimatedEventLocation `json:"estimatedEventLocation,omitempty"`
	AlertTopics            []AlertTopic            `json:"alertTopics,omitempty"`
	LiveBrief              []LiveBrief             `json:"liveBrief,omitempty"`
	DataminrAlertURL       string                  `json:"dataminrAlertUrl,omitempty"`
	AlertReferenceTerms    []AlertReferenceTerms   `json:"alertReferenceTerms,omitempty"`
	ListsMatched           []ListsMatched          `json:"listsMatched,omitempty"`
	LinkedAlerts           []LinkedAlerts          `json:"linkedAlerts,omitempty"`
	TermsOfUse             string                  `json:"termsOfUse,omitempty"`
}

// AlertType represents the criticality level of an alert
type AlertType struct {
	Name string `json:"name"`
}

// AlertTopic represents topic information for an alert
type AlertTopic struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// LinkedAlerts indicates multiple alerts linked as part of a broader event
type LinkedAlerts struct {
	Count         int    `json:"count"`
	ParentAlertID string `json:"parentAlertId"`
}

// SubHeadline provides additional context about an alert
type SubHeadline struct {
	Title   string   `json:"title"`
	Content []string `json:"content"`
}

// PublicPost contains the URL of the public post
type PublicPost struct {
	Href string `json:"href"`
}

// EstimatedEventLocation represents the estimated location of the event
type EstimatedEventLocation struct {
	Name              string    `json:"name"`
	Coordinates       []float64 `json:"coordinates"`
	ProbabilityRadius float64   `json:"probabilityRadius"`
	MGRS              string    `json:"MGRS"`
}

// LiveBrief contains AI-generated summary of the holistic event
type LiveBrief struct {
	Summary   string `json:"summary"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}

// AlertReferenceTerms contains keywords related to the alert
type AlertReferenceTerms struct {
	Text string `json:"text"`
}

// ListsMatched represents alert lists that delivered the alert
type ListsMatched struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	TopicIDs []string `json:"topicIds"`
}
