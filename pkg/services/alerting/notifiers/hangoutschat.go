package notifiers

import (
	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/log"
	m "github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/alerting"
)

func init() {
	alerting.RegisterNotifier(&alerting.NotifierPlugin{
		Type:        "hangoutschat",
		Name:        "hangoutschat",
		Description: "Sends Hangouts Chat request to a URL",
		Factory:     NewHangoutsChatNotifier,
		OptionsTemplate: `
      <h3 class="page-heading">Hangouts Chat settings</h3>
      <div class="gf-form">
        <span class="gf-form-label width-10">Url</span>
        <input type="text" required class="gf-form-input max-width-26" ng-model="ctrl.model.settings.url"></input>
      </div>
    `,
	})

}

func NewHangoutsChatNotifier(model *m.AlertNotification) (alerting.Notifier, error) {
	url := model.Settings.Get("url").MustString()
	if url == "" {
		return nil, alerting.ValidationError{Reason: "Could not find url property in settings"}
	}

	return &HangoutsChatNotifier{
		NotifierBase: NewNotifierBase(model),
		Url:          url,
		log:          log.New("alerting.notifier.hangoutschat"),
	}, nil
}

type HangoutsChatNotifier struct {
	NotifierBase
	Url        string
	log        log.Logger
}

func (this *HangoutsChatNotifier) Notify(evalContext *alerting.EvalContext) error {
	this.log.Info("Sending hangouts chat")

	bodyJSON := simplejson.New()
	if evalContext.ImagePublicUrl != "" {
		bodyJSON.Set("imageUrl", evalContext.ImagePublicUrl)
	}

	if evalContext.Rule.Message != "" {
		bodyJSON.Set("text", evalContext.Rule.Message)
	}

	body, _ := bodyJSON.MarshalJSON()

	cmd := &m.SendWebhookSync{
		Url:        this.Url,
		Body:       string(body),
		HttpMethod: "POST",
	}

	if err := bus.DispatchCtx(evalContext.Ctx, cmd); err != nil {
		this.log.Error("Failed to send hangouts chat", "error", err, "hangoutschat", this.Name)
		return err
	}

	return nil
}
