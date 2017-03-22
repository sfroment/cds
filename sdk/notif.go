package sdk

import (
	"encoding/json"
	"strings"
)

//UserNotificationSettingsType of notification
type UserNotificationSettingsType string

//const
const (
	EmailUserNotification  UserNotificationSettingsType = "email"
	JabberUserNotification UserNotificationSettingsType = "jabber"
)

//UserNotificationEventType always/never/change
type UserNotificationEventType string

//const
const (
	UserNotificationAlways UserNotificationEventType = "always"
	UserNotificationNever  UserNotificationEventType = "never"
	UserNotificationChange UserNotificationEventType = "change"
)

//UserNotificationDefaultSettings are maps of default values
var UserNotificationDefaultSettings = map[UserNotificationSettingsType]map[string]string{
	EmailUserNotification: map[string]string{
		"subject": "{{.cds.project}}/{{.cds.application}} {{.cds.pipeline}} {{.cds.environment}}#{{.cds.version}} {{.cds.status}}",
		"body": `Project : {{.cds.project}}
Application : {{.cds.application}}
Pipeline : {{.cds.pipeline}}/{{.cds.environment}}#{{.cds.version}}
Status : {{.cds.status}}
Details : {{.cds.buildURL}}
Triggered by : {{.cds.triggered_by.username}}
Branch : {{.git.branch}}`,
	},
	JabberUserNotification: map[string]string{
		"subject": "{{.cds.project}}/{{.cds.application}} {{.cds.pipeline}} {{.cds.environment}}#{{.cds.version}} {{.cds.status}}",
		"body": `Project : {{.cds.project}}
Application : {{.cds.application}}
Pipeline : {{.cds.pipeline}}/{{.cds.environment}}#{{.cds.version}}
Status : {{.cds.status}}
Details : {{.cds.buildURL}}
Triggered by : {{.cds.triggered_by.username}}
Branch : {{.git.branch}}`,
	},
}

// UserNotification is a settings on application_pipeline/env
// to trigger notification on pipeline event
type UserNotification struct {
	ApplicationPipelineID int64                                                     `json:"application_pipeline_id"`
	Pipeline              Pipeline                                                  `json:"pipeline"`
	Environment           Environment                                               `json:"environment"`
	Notifications         map[UserNotificationSettingsType]UserNotificationSettings `json:"notifications"`
}

//UnmarshalJSON parses the JSON-encoded data and stores the result in n
func (n *UserNotification) UnmarshalJSON(b []byte) error {
	notif, err := parseUserNotification(b)
	if err != nil {
		return err
	}
	*n = *notif
	return nil
}

// UserNotificationSettings are common settings
type UserNotificationSettings interface {
	Success() UserNotificationEventType
	Failure() UserNotificationEventType
	Start() bool
	JSON() string
	Config() map[string]interface{}
}

// JabberEmailUserNotificationSettings are jabber or email settings
type JabberEmailUserNotificationSettings struct {
	OnSuccess    UserNotificationEventType `json:"on_success"`     //default value is 'change'
	OnFailure    UserNotificationEventType `json:"on_failure"`     //default value is 'always'
	OnStart      bool                      `json:"on_start"`       //default value is 'never'
	SendToGroups bool                      `json:"send_to_groups"` //default value is 'false'
	SendToAuthor bool                      `json:"send_to_author"` //default value is 'true'
	Recipients   []string                  `json:"recipients"`
	Template     UserNotificationTemplate  `json:"template"`
}

//Success returns always/never/change
func (n *JabberEmailUserNotificationSettings) Success() UserNotificationEventType {
	return n.OnSuccess
}

//Failure returns always/never/change
func (n *JabberEmailUserNotificationSettings) Failure() UserNotificationEventType {
	return n.OnFailure
}

//Start returns always/never/change
func (n *JabberEmailUserNotificationSettings) Start() bool {
	return n.OnStart
}

//JSON returns json as string
func (n *JabberEmailUserNotificationSettings) JSON() string {
	b, _ := json.Marshal(n)
	return string(b)
}

//Config return a map of user notification settings
func (n *JabberEmailUserNotificationSettings) Config() map[string]interface{} {
	e := map[string]interface{}{}

	if n.OnSuccess != UserNotificationChange {
		if string(n.OnSuccess) == "" {
			e["on_success"] = string(UserNotificationNever)
		} else {
			e["on_success"] = string(n.OnSuccess)
		}
	}

	if n.OnFailure != UserNotificationAlways {
		if string(n.OnFailure) == "" {
			e["on_failure"] = string(UserNotificationNever)
		} else {
			e["on_failure"] = string(n.OnFailure)
		}
	}

	if n.OnStart {
		e["on_start"] = n.OnStart
	}

	if !n.SendToAuthor {
		e["send_to_author"] = n.SendToAuthor
	}

	if n.SendToGroups {
		e["send_to_groups"] = n.SendToGroups
	}

	if n.Recipients != nil {
		e["recipients"] = strings.Join(n.Recipients, ",")
	}

	d := UserNotificationDefaultSettings[JabberUserNotification]

	if n.Template.Body != d["body"] {
		e["body"] = n.Template.Body
	}

	if n.Template.Body != d["subject"] {
		e["subject"] = n.Template.Subject
	}

	return e
}

// UserNotificationTemplate is the notification content
type UserNotificationTemplate struct {
	Subject string `json:"subject,omitempty"`
	Body    string `json:"body,omitempty"`
}

//userNotificationInput is a way to parse notification
type userNotificationInput struct {
	Notifications         map[string]interface{} `json:"notifications"`
	ApplicationPipelineID int64                  `json:"application_pipeline_id"`
	Environment           Environment            `json:"environment"`
	Pipeline              Pipeline               `json:"pipeline"`
}

//ParseUserNotification transform jsons to UserNotificationSettings map
func parseUserNotification(body []byte) (*UserNotification, error) {
	var input = &userNotificationInput{}
	if err := json.Unmarshal(body, &input); err != nil {
		return nil, err
	}
	settingsBody, err := json.Marshal(input.Notifications)
	if err != nil {
		return nil, err
	}

	var notif1 = &UserNotification{
		ApplicationPipelineID: input.ApplicationPipelineID,
		Environment:           input.Environment,
		Pipeline:              input.Pipeline,
	}

	var errParse error
	notif1.Notifications, errParse = ParseUserNotificationSettings(settingsBody)
	return notif1, errParse
}

//ParseUserNotificationSettings transforms json to UserNotificationSettings map
func ParseUserNotificationSettings(settings []byte) (map[UserNotificationSettingsType]UserNotificationSettings, error) {
	mapSettings := map[string]interface{}{}
	if err := json.Unmarshal(settings, &mapSettings); err != nil {
		return nil, err
	}

	notifications := map[UserNotificationSettingsType]UserNotificationSettings{}

	for k, v := range mapSettings {
		switch k {
		case string(EmailUserNotification), string(JabberUserNotification):
			if v != nil {
				var x JabberEmailUserNotificationSettings
				tmp, err := json.Marshal(v)
				if err != nil {
					return nil, ErrParseUserNotification
				}
				if err := json.Unmarshal(tmp, &x); err != nil {
					return nil, ErrParseUserNotification
				}
				notifications[UserNotificationSettingsType(k)] = &x
			}
		default:
			return nil, ErrNotSupportedUserNotification
		}
	}

	return notifications, nil
}
