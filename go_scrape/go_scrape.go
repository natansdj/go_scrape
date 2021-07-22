package go_scrape

import (
	"net/http"
	"net/url"
	"sync"

	"github.com/natansdj/go_scrape/config"
	"github.com/natansdj/go_scrape/logx"
	"github.com/natansdj/go_scrape/queue"

	"github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// D provide string array
type D map[string]interface{}

const (
	// ApnsPriorityLow will tell APNs to send the push message at a time that takes
	// into account power considerations for the device. Notifications with this
	// priority might be grouped and delivered in bursts. They are throttled, and
	// in some cases are not delivered.
	ApnsPriorityLow = 5

	// ApnsPriorityHigh will tell APNs to send the push message immediately.
	// Notifications with this priority must trigger an alert, sound, or badge on
	// the target device. It is an error to use this priority for a push
	// notification that contains only the content-available key.
	ApnsPriorityHigh = 10
)

// Alert is APNs payload
type Alert struct {
	Action          string   `json:"action,omitempty"`
	ActionLocKey    string   `json:"action-loc-key,omitempty"`
	Body            string   `json:"body,omitempty"`
	LaunchImage     string   `json:"launch-image,omitempty"`
	LocArgs         []string `json:"loc-args,omitempty"`
	LocKey          string   `json:"loc-key,omitempty"`
	Title           string   `json:"title,omitempty"`
	Subtitle        string   `json:"subtitle,omitempty"`
	TitleLocArgs    []string `json:"title-loc-args,omitempty"`
	TitleLocKey     string   `json:"title-loc-key,omitempty"`
	SummaryArg      string   `json:"summary-arg,omitempty"`
	SummaryArgCount int      `json:"summary-arg-count,omitempty"`
}

// RequestPush support multiple notification request.
type RequestPush struct {
	Notifications []PushNotification `json:"notifications" binding:"required"`
}

// PushNotification is single notification request
type PushNotification struct {
	Wg  *sync.WaitGroup
	Log *[]logx.LogPushEntry
	Cfg config.ConfYaml

	// Common
	ID               string      `json:"notif_id,omitempty"`
	Tokens           []string    `json:"tokens" binding:"required"`
	Platform         int         `json:"platform" binding:"required"`
	Message          string      `json:"message,omitempty"`
	Title            string      `json:"title,omitempty"`
	Image            string      `json:"image,omitempty"`
	Priority         string      `json:"priority,omitempty"`
	ContentAvailable bool        `json:"content_available,omitempty"`
	MutableContent   bool        `json:"mutable_content,omitempty"`
	Sound            interface{} `json:"sound,omitempty"`
	Data             D           `json:"data,omitempty"`
	Retry            int         `json:"retry,omitempty"`

	// Android
	APIKey                string `json:"api_key,omitempty"`
	To                    string `json:"to,omitempty"`
	CollapseKey           string `json:"collapse_key,omitempty"`
	DelayWhileIdle        bool   `json:"delay_while_idle,omitempty"`
	TimeToLive            *uint  `json:"time_to_live,omitempty"`
	RestrictedPackageName string `json:"restricted_package_name,omitempty"`
	DryRun                bool   `json:"dry_run,omitempty"`
	Condition             string `json:"condition,omitempty"`

	// Huawei
	AppID         string `json:"app_id,omitempty"`
	AppSecret     string `json:"app_secret,omitempty"`
	BiTag         string `json:"bi_tag,omitempty"`
	FastAppTarget int    `json:"fast_app_target,omitempty"`

	// iOS
	Expiration  *int64   `json:"expiration,omitempty"`
	ApnsID      string   `json:"apns_id,omitempty"`
	CollapseID  string   `json:"collapse_id,omitempty"`
	Topic       string   `json:"topic,omitempty"`
	PushType    string   `json:"push_type,omitempty"`
	Badge       *int     `json:"badge,omitempty"`
	Category    string   `json:"category,omitempty"`
	ThreadID    string   `json:"thread-id,omitempty"`
	URLArgs     []string `json:"url-args,omitempty"`
	Alert       Alert    `json:"alert,omitempty"`
	Production  bool     `json:"production,omitempty"`
	Development bool     `json:"development,omitempty"`
	SoundName   string   `json:"name,omitempty"`
	SoundVolume float32  `json:"volume,omitempty"`
	Apns        D        `json:"apns,omitempty"`
}

// WaitDone decrements the WaitGroup counter.
func (p *PushNotification) WaitDone() {
	if p.Wg != nil {
		p.Wg.Done()
	}
}

// AddWaitCount increments the WaitGroup counter.
func (p *PushNotification) AddWaitCount() {
	if p.Wg != nil {
		p.Wg.Add(1)
	}
}

// AddLog record fail log of notification
func (p *PushNotification) AddLog(log logx.LogPushEntry) {
	if p.Log != nil {
		*p.Log = append(*p.Log, log)
	}
}

// Bytes for queue message
func (p *PushNotification) Bytes() []byte {
	b, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	return b
}

// SetProxy only working for FCM server.
func SetProxy(proxy string) error {
	proxyURL, err := url.ParseRequestURI(proxy)
	if err != nil {
		return err
	}

	http.DefaultTransport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	logx.LogAccess.Debug("Set http proxy as " + proxy)

	return nil
}

// SendNotification send notification
func SendNotification(req queue.QueuedMessage) {
	v, _ := req.(*PushNotification)

	defer func() {
		v.WaitDone()
	}()

	switch v.Platform {
	default:
		PushToPlatform(*v)
	}
}

// PushToPlatform provide send notification to Android server.
func PushToPlatform(req PushNotification) {
	logx.LogAccess.Debug("Start push notification to platform", req)
}
