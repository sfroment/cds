package exportentities

import (
	"fmt"

	"strings"

	"github.com/ovh/cds/sdk"
)

// Application represents exported sdk.Application
type Application struct {
	Name              string                         `json:"name" yaml:"name"`
	RepositoryManager string                         `json:"repo_manager,omitempty" yaml:"repo_manager,omitempty"`
	RepositoryName    string                         `json:"repo_name,omitempty" yaml:"repo_name,omitempty"`
	Permissions       map[string]int                 `json:"permissions,omitempty" yaml:"permissions,omitempty"`
	Variables         map[string]VariableValue       `json:"variables,omitempty" yaml:"variables,omitempty"`
	Pipelines         map[string]ApplicationPipeline `json:"pipelines,omitempty" yaml:"pipelines,omitempty"`
}

// ApplicationPipeline represents exported sdk.ApplicationPipeline
type ApplicationPipeline struct {
	Parameters map[string]VariableValue                `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Triggers   map[string][]ApplicationPipelineTrigger `json:"triggers,omitempty" yaml:"triggers,omitempty"`
	Options    []ApplicationPipelineOptions            `json:"options,omitempty" yaml:"options,omitempty"`
}

// ApplicationPipelineOptions represents presence of hooks, pollers, notifications and scheduler for an tuple application pipeline environment
type ApplicationPipelineOptions struct {
	Environment   *string                                    `json:"environment,omitempty" yaml:"environment,omitempty"`
	Hook          *bool                                      `json:"hook,omitempty" yaml:"hook,omitempty"`
	Polling       *bool                                      `json:"polling,omitempty" yaml:"polling,omitempty"`
	Notifications map[string]ApplicationPipelineNotification `json:"notifications,omitempty" yaml:"notifications,omitempty"`
	Schedulers    []ApplicationPipelineScheduler             `json:"schedulers,omitempty" yaml:"schedulers,omitempty"`
}

// ApplicationPipelineScheduler represents exported sdk.PipelineScheduler
type ApplicationPipelineScheduler struct {
	CronExpr   string                   `json:"cron_expr" yaml:"cron_expr"`
	Parameters map[string]VariableValue `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

// ApplicationPipelineNotification represents exported notification
type ApplicationPipelineNotification map[string]interface{}

// ApplicationPipelineTrigger represents an exported pipeline trigger
type ApplicationPipelineTrigger struct {
	ProjectKey      *string                  `json:"project_key,omitempty" yaml:"project_key,omitempty"`
	ApplicationName *string                  `json:"application_name,omitempty" yaml:"application_name,omitempty"`
	FromEnvironment *string                  `json:"from_environment,omitempty" yaml:"from_environment,omitempty"`
	ToEnvironment   *string                  `json:"to_environment,omitempty" yaml:"environment,omitempty"`
	Manual          bool                     `json:"manual" yaml:"manual"`
	Conditions      []Condition              `json:"conditions,omitempty" yaml:"conditions,omitempty"`
	Parameters      map[string]VariableValue `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

// Condition represents sdk.Prerequisite
type Condition struct {
	Variable string `json:"variable" yaml:"variable"`
	Expected string `json:"expected" yaml:"expected"`
}

// NewApplication instanciance an exportable application from an sdk.Application
func NewApplication(app *sdk.Application) (a *Application) {
	a = new(Application)
	a.Name = app.Name

	if app.RepositoriesManager != nil {
		a.RepositoryManager = app.RepositoriesManager.Name
		a.RepositoryName = app.RepositoryFullname
	}

	a.Variables = make(map[string]VariableValue, len(app.Variable))
	for _, v := range app.Variable {
		a.Variables[v.Name] = VariableValue{
			Type:  string(v.Type),
			Value: v.Value,
		}
	}
	a.Permissions = make(map[string]int, len(app.ApplicationGroups))
	for _, p := range app.ApplicationGroups {
		a.Permissions[p.Group.Name] = p.Permission
	}

	a.Pipelines = make(map[string]ApplicationPipeline, len(app.Pipelines))
	for _, ap := range app.Pipelines {
		pip := ApplicationPipeline{}

		pip.Parameters = make(map[string]VariableValue, len(ap.Parameters))
		for _, param := range ap.Parameters {
			pip.Parameters[param.Name] = VariableValue{
				Type:  string(param.Type),
				Value: param.Value,
			}
		}

		pip.Triggers = map[string][]ApplicationPipelineTrigger{}
		for i := range ap.Triggers {
			t := &ap.Triggers[i]
			fmt.Println("Trigger : ", t.SrcPipeline.Name, t.SrcApplication.Name, t.SrcEnvironment.Name, t.DestPipeline.Name, t.DestApplication.Name, t.DestEnvironment.Name)
			if ap.Pipeline.Name != t.SrcPipeline.Name {
				continue
			}
			if a.Name != t.SrcApplication.Name {
				continue
			}

			//Compute trigger Prerequisites
			c := make([]Condition, len(t.Prerequisites))
			var i int
			for _, pr := range t.Prerequisites {
				c[i] = Condition{
					Variable: pr.Parameter,
					Expected: pr.ExpectedValue,
				}
			}

			//Compute trigger parameters
			p := map[string]VariableValue{}
			for _, param := range t.Parameters {
				p[param.Name] = VariableValue{
					Type:  param.Type,
					Value: param.Value,
				}
			}

			var srcEnv, destEnv, pKey, appName *string
			if t.SrcEnvironment.Name != sdk.DefaultEnv.Name {
				srcEnv = &t.SrcEnvironment.Name
			}
			if t.DestEnvironment.Name != sdk.DefaultEnv.Name {
				destEnv = &t.DestEnvironment.Name
			}
			if t.DestProject.Key != app.ProjectKey {
				pKey = &t.DestProject.Key
			}
			if t.DestApplication.Name != a.Name {
				appName = &t.DestApplication.Name
			}
			ap := ApplicationPipelineTrigger{
				ProjectKey:      pKey,
				ApplicationName: appName,
				ToEnvironment:   destEnv,
				FromEnvironment: srcEnv,
				Conditions:      c,
				Parameters:      p,
			}
			ap.Manual = t.Manual
			if pip.Triggers[t.DestPipeline.Name] == nil {
				pip.Triggers[t.DestPipeline.Name] = []ApplicationPipelineTrigger{}
			}
			pip.Triggers[t.DestPipeline.Name] = append(pip.Triggers[t.DestPipeline.Name], ap)
		}

		mapEnvOpts := map[string]*ApplicationPipelineOptions{}
		//Hooks
		for _, h := range app.Hooks {
			if h.Enabled && h.Pipeline.Name == ap.Pipeline.Name {
				if _, ok := mapEnvOpts[sdk.DefaultEnv.Name]; !ok {
					mapEnvOpts[sdk.DefaultEnv.Name] = &ApplicationPipelineOptions{}
				}
				o := mapEnvOpts[sdk.DefaultEnv.Name]
				if h.Enabled {
					var ok = true
					o.Hook = &ok
				}
			}
		}

		//Pollers
		for _, p := range app.RepositoryPollers {
			if p.Enabled && p.Pipeline.Name == ap.Pipeline.Name {
				if _, ok := mapEnvOpts[sdk.DefaultEnv.Name]; !ok {
					mapEnvOpts[sdk.DefaultEnv.Name] = &ApplicationPipelineOptions{}
				}
				o := mapEnvOpts[sdk.DefaultEnv.Name]
				var ok = true
				o.Polling = &ok
			}

		}

		//Notifications
		for _, n := range app.Notifications {
			if ap.Pipeline.Name == n.Pipeline.Name {
				if _, ok := mapEnvOpts[n.Environment.Name]; !ok {
					mapEnvOpts[n.Environment.Name] = &ApplicationPipelineOptions{}
				}
				o := mapEnvOpts[n.Environment.Name]
				for t, n := range n.Notifications {
					if o.Notifications == nil {
						o.Notifications = make(map[string]ApplicationPipelineNotification)
					}
					o.Notifications[string(t)] = n.Config()
				}
			}
		}

		//Schedulers
		for _, s := range app.Schedulers {
			if ap.Pipeline.ID == s.PipelineID {
				if _, ok := mapEnvOpts[s.EnvironmentName]; !ok {
					mapEnvOpts[s.EnvironmentName] = &ApplicationPipelineOptions{}
				}
				o := mapEnvOpts[s.EnvironmentName]
				if o.Schedulers == nil {
					o.Schedulers = []ApplicationPipelineScheduler{}
				}
				aps := ApplicationPipelineScheduler{
					CronExpr: s.Crontab,
				}
				aps.Parameters = make(map[string]VariableValue, len(s.Args))
				for _, p := range s.Args {
					aps.Parameters[p.Name] = VariableValue{Type: string(p.Type), Value: p.Value}
				}
				o.Schedulers = append(o.Schedulers, aps)
			}
		}

		//Compute all
		pip.Options = make([]ApplicationPipelineOptions, len(mapEnvOpts))
		var i int
		for k, v := range mapEnvOpts {
			if k != sdk.DefaultEnv.Name {
				s := k
				pip.Options[i].Environment = &s
			}
			if v.Hook != nil {
				pip.Options[i].Hook = v.Hook
			}
			if v.Polling != nil {
				pip.Options[i].Polling = v.Polling
			}
			pip.Options[i].Notifications = v.Notifications
			pip.Options[i].Schedulers = v.Schedulers

			i++
		}

		fmt.Println(ap.Pipeline.Name, len(pip.Options), len(pip.Parameters), len(pip.Triggers))
		var ignore bool
		if len(pip.Options) == 0 && len(pip.Parameters) == 0 && len(pip.Triggers) == 0 {
			for _, v := range a.Pipelines {
				for b := range v.Triggers {
					if b == ap.Pipeline.Name {
						ignore = true
						break
					}
				}
			}
		}
		if !ignore {
			a.Pipelines[ap.Pipeline.Name] = pip
		}
	}

	return
}

//Application returns a sdk.Application
func (a *Application) Application() (*sdk.Application, error) {
	app := &sdk.Application{
		Name: a.Name,
	}

	if a.RepositoryManager != "" {
		app.RepositoriesManager = &sdk.RepositoriesManager{
			Name: a.RepositoryManager,
		}
	}

	if a.RepositoryName != "" {
		app.RepositoryFullname = a.RepositoryName
	}

	for k, v := range a.Permissions {
		app.ApplicationGroups = append(app.ApplicationGroups, sdk.GroupPermission{
			Group:      sdk.Group{Name: k},
			Permission: v,
		})
	}

	for k, v := range a.Variables {
		app.Variable = append(app.Variable, sdk.Variable{
			Name:  k,
			Type:  v.Type,
			Value: v.Value,
		})
	}

	for pipelineName, applicationPipeline := range a.Pipelines {
		ap := sdk.ApplicationPipeline{
			Pipeline: sdk.Pipeline{Name: pipelineName},
		}

		for k, v := range applicationPipeline.Parameters {
			ap.Parameters = append(ap.Parameters, sdk.Parameter{
				Name:  k,
				Type:  v.Type,
				Value: v.Value,
			})
		}

		for d, triggers := range applicationPipeline.Triggers {
			for _, t := range triggers {
				trig := sdk.PipelineTrigger{
					DestPipeline: sdk.Pipeline{Name: d},
					SrcPipeline:  sdk.Pipeline{Name: pipelineName},
				}

				if t.FromEnvironment != nil {
					trig.SrcEnvironment = sdk.Environment{
						Name: *t.FromEnvironment,
					}
				}

				if t.ToEnvironment != nil {
					trig.DestEnvironment = sdk.Environment{
						Name: *t.ToEnvironment,
					}
				}

				if t.ApplicationName != nil {
					trig.DestApplication = sdk.Application{Name: *t.ApplicationName}
				} else {
					trig.DestApplication = sdk.Application{Name: a.Name}
				}

				if t.ProjectKey != nil {
					trig.DestProject = sdk.Project{Key: *t.ProjectKey}
				}

				for _, c := range t.Conditions {
					trig.Prerequisites = append(trig.Prerequisites, sdk.Prerequisite{
						Parameter:     c.Variable,
						ExpectedValue: c.Expected,
					})
				}

				ap.Triggers = append(ap.Triggers, trig)
			}

			for k, v := range applicationPipeline.Parameters {
				ap.Parameters = append(ap.Parameters, sdk.Parameter{
					Name:  k,
					Type:  v.Type,
					Value: v.Value,
				})
			}

			for _, o := range applicationPipeline.Options {
				env := sdk.DefaultEnv.Name
				if o.Environment != nil {
					env = *o.Environment
				}

				//Compute hooks
				if o.Hook != nil && *o.Hook {
					app.Hooks = append(app.Hooks, sdk.Hook{
						Enabled: true,
						Pipeline: sdk.Pipeline{
							Name: pipelineName,
						},
					})
				}

				//Compute pollers
				if o.Polling != nil && *o.Polling {
					app.RepositoryPollers = append(app.RepositoryPollers, sdk.RepositoryPoller{
						Pipeline: sdk.Pipeline{
							Name: pipelineName,
						},
					})
				}

				//Compute notifications
				notifs := map[sdk.UserNotificationSettingsType]sdk.UserNotificationSettings{}
				for k, v := range o.Notifications {
					switch k {
					case string(sdk.JabberUserNotification), string(sdk.EmailUserNotification):
						notif := sdk.JabberEmailUserNotificationSettings{}
						if v["on_success"] == nil {
							notif.OnSuccess = sdk.UserNotificationNever
						} else {
							notif.OnSuccess = sdk.UserNotificationEventType(v["on_success"].(string))
						}
						if v["on_failure"] == nil {
							notif.OnFailure = sdk.UserNotificationAlways
						} else {
							str, ok := v["on_failure"].(string)
							if !ok {
								return nil, fmt.Errorf("Unrecogized notification.on_failure (%v) option on pipeline %s", v["on_failure"], pipelineName)
							}
							notif.OnFailure = sdk.UserNotificationEventType(str)
						}
						if v["on_start"] == nil {
							notif.OnStart = false
						} else {
							var ok bool
							notif.OnStart, ok = v["on_start"].(bool)
							if !ok {
								return nil, fmt.Errorf("Unrecogized notification.on_start (%v) option on pipeline %s", v["on_start"], pipelineName)
							}
						}
						str, ok := v["recipients"].(string)
						if !ok {
							return nil, fmt.Errorf("Unrecogized notification.recipients (%v) option on pipeline %s", v["recipients"], pipelineName)
						}
						notif.Recipients = strings.Split(str, ",")

						//send_to_author
						if v["send_to_author"] == nil {
							notif.SendToAuthor = false
						} else {
							var ok bool
							notif.SendToAuthor, ok = v["send_to_author"].(bool)
							if !ok {
								return nil, fmt.Errorf("Unrecogized notification.send_to_author (%v) option on pipeline %s", v["send_to_author"], pipelineName)
							}
						}

						//send_to_groups
						if v["send_to_groups"] == nil {
							notif.SendToGroups = true
						} else {
							var ok bool
							notif.SendToGroups, ok = v["send_to_groups"].(bool)
							if !ok {
								return nil, fmt.Errorf("Unrecogized notification.send_to_groups (%v) option on pipeline %s", v["send_to_groups"], pipelineName)
							}
						}

						d := sdk.UserNotificationDefaultSettings[sdk.UserNotificationSettingsType(k)]

						//body
						str, ok = v["body"].(string)
						if !ok {
							return nil, fmt.Errorf("Unrecogized notification.body (%v) option on pipeline %s", v["body"], pipelineName)
						}
						if str == "" {
							str = d["body"]
						}
						notif.Template.Body = str

						//subject
						str, ok = v["subject"].(string)
						if !ok {
							return nil, fmt.Errorf("Unrecogized notification.subject (%v) option on pipeline %s", d["subject"], pipelineName)
						}
						if str == "" {
							str = d["subject"]
						}
						notif.Template.Subject = str

						notifs[sdk.UserNotificationSettingsType(k)] = &notif
					}
				}

				app.Notifications = append(app.Notifications, sdk.UserNotification{
					Environment:   sdk.Environment{Name: env},
					Pipeline:      sdk.Pipeline{Name: pipelineName},
					Notifications: notifs,
				})

				//Compute schedulers
				for _, s := range o.Schedulers {
					sched := sdk.PipelineScheduler{
						EnvironmentName: env,
						PipelineName:    pipelineName,
						Crontab:         s.CronExpr,
					}

					for k, v := range s.Parameters {
						sched.Args = append(sched.Args, sdk.Parameter{
							Name:  k,
							Type:  v.Type,
							Value: v.Value,
						})
					}

					app.Schedulers = append(app.Schedulers, sched)
				}
			}
		}

		app.Pipelines = append(app.Pipelines, ap)
	}

	//Browse all pipelines triggers to add triggered pipeline wich has been ignored in application pipelines
	var ignoredPipelines []sdk.ApplicationPipeline
	for _, p := range app.Pipelines {
		for _, t := range p.Triggers {
			if t.DestApplication.Name == app.Name {
				var ignored = true
				for _, p2 := range app.Pipelines {
					if t.DestPipeline.Name == p2.Pipeline.Name {
						ignored = false
						break
					}
				}
				if ignored {
					var foundIgnoredPipelines bool
					for i := range ignoredPipelines {
						if t.DestPipeline.Name == ignoredPipelines[i].Pipeline.Name {
							foundIgnoredPipelines = true
							break
						}
					}
					if !foundIgnoredPipelines {
						ignoredPipelines = append(ignoredPipelines, sdk.ApplicationPipeline{
							Pipeline: sdk.Pipeline{Name: t.DestPipeline.Name},
						})
					}
				}
			}

		}
	}
	app.Pipelines = append(app.Pipelines, ignoredPipelines...)

	return app, nil
}
