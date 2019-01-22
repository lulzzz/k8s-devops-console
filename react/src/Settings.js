import React from 'react';
import $ from 'jquery';

import BaseComponent from './BaseComponent';
import Spinner from "./Spinner";

class Settings extends BaseComponent {
    constructor(props) {
        super(props);

        this.state = {
            isStartupConfig: true,
            isStartupSettings: true,
            globalError: "",
            globalMessage: "",
            buttonState: "",

            config: {
                User: {
                    Username: '',
                },
                Teams: [],
                NamespaceEnvironments: [],
                Quota: {}
            },

            requestRunning: false,

            personalMessage: "",

            personal: {
                SshPubKey: "",
            },

            team: {

            },
        };
    }

    componentDidMount() {
        this.loadConfig();
        this.loadSettings();
    }

    loadConfig() {
        $.get({
            url: '/api/_app/config'
        }).done((jqxhr) => {
            if (jqxhr) {
                if (!jqxhr.Teams) {
                    jqxhr.Teams = [];
                }

                if (!jqxhr.NamespaceEnvironments) {
                    jqxhr.NamespaceEnvironments = [];
                }

                this.setState({
                    config: jqxhr,
                    isStartupConfig: false
                });
            }
        });
    }

    loadSettings() {
        $.get({
            url: '/api/settings'
        }).done((jqxhr) => {
            if (jqxhr) {
                var state = this.state;
                state.isStartupSettings = false;

                if (jqxhr.personal) {
                    state.personal = jqxhr.personal;
                }

                if (jqxhr.team) {
                    state.team = jqxhr.team;
                }

                this.setState(state);
            }
        });
    }

    handlePersonalInputChange(name, event) {
        var state = this.state.personal;
        state[name] = event.target.value;
        this.setState(state);
    }

    handleTeamInputChange(team, name, event) {
        var state = this.state.team;

        if (!state[team]) {
            state[team] = {};
        }

        state[team][name] = event.target.value;
        this.setState(state);
    }

    stateUpdateButton() {
        let state = "";

        if (this.state.requestRunning) {
            state = "disabled";
        }

        return state
    }

    updatePersonalSettings(e) {
        e.preventDefault();
        e.stopPropagation();

        this.setState({
            requestRunning: true,
            globalError: "",
            personalMessage: "",
            teamMessage: {}
        });

        let jqxhr = $.ajax({
            type: 'POST',
            url: "/api/settings/personal",
            data: {
                config: this.state.personal
            }
        }).done((jqxhr) => {
            this.setState({
                personalMessage: "Personal settings updated",
            });
        }).always(() => {
            this.setState({
                requestRunning: false,
            });
        });

        this.handleXhr(jqxhr);
    }

    updateTeamSettings(team, e) {
        e.preventDefault();
        e.stopPropagation();


        this.setState({
            requestRunning: true,
            globalError: "",
            personalMessage: "",
            teamMessage: {}
        });

        let jqxhr = $.ajax({
            type: 'POST',
            url: "/api/settings/team",
            data: {
                team: team,
                config: this.getTeamConfig(team)
            }
        }).done((jqxhr) => {
            var state = {
                teamMessage: {}
            };
            state.teamMessage[team] = "Team " + team + " settings updated";
            this.setState(state);
        }).always(() => {
            this.setState({
                requestRunning: false,
            });
        });

        this.handleXhr(jqxhr);
    }

    getTeamConfig(team) {
        var ret = {};

        if (this.state.team && this.state.team[team]) {
            ret = this.state.team[team];
        }

        return ret;
    }
    getTeamConfigItem(team, name) {
        var ret = "";

        if (this.state.team && this.state.team[team] && this.state.team[team][name]) {
            ret = this.state.team[team][name];
        }

        return ret;
    }

    isStartup() {
        return this.state.isStartupConfig || this.state.isStartupSettings
    }

    getTeamMessage(team) {
        if (this.state.teamMessage && this.state.teamMessage[team]) {
            return this.state.teamMessage[team];
        }
        return false
    }

    render() {
        if ((this.state.isStartupConfig || this.state.isStartupSettings) && this.state.globalError) {
            return (
                <div className="alert alert-danger">{this.state.globalError}</div>
            )
        }

        return (
            <div>
                <Spinner active={this.isStartup()}/>
                <div>
                    <div className={this.state.globalError === '' ? null : 'alert alert-danger'}>{this.state.globalError}</div>
                    <div className={this.state.globalMessage === '' ? 'alert alert-success invisible' : 'alert alert-success'}>{this.state.globalMessage}</div>
                </div>

                <h2>Personal settings</h2>
                <div>
                    <div className={this.state.personalMessage === '' ? 'alert alert-success invisible' : 'alert alert-success'}>{this.state.personalMessage}</div>
                </div>
                <form method="post">
                    <div className="form-group">
                        <label htmlFor="inputNsApp" className="inputRg">SSH Public Key</label>
                        <input type="text" name="personalSshKey" id="personalSshKey" className="form-control" placeholder="SSH Public Key" value={this.state.personal.SshPubKey} onChange={this.handlePersonalInputChange.bind(this,"SshPubKey")} />
                    </div>
                    <div className="toolbox">
                        <button type="submit" className="btn btn-primary bnt-k8s-namespace-create" disabled={this.stateUpdateButton()} onClick={this.updatePersonalSettings.bind(this)}>Save</button>
                    </div>
                </form>


                {this.state.config.Teams.map((row, value) =>
                    <div>
                        <h2>Team {row.Name} settings</h2>
                        <div>
                            <div className={this.getTeamMessage(row.Name) === false ? 'alert alert-success invisible' : 'alert alert-success'}>{this.getTeamMessage(row.Name)}</div>
                        </div>
                        <form method="post">
                            <div className="form-group">
                                <label htmlFor="inputNsApp" className="inputRg">Alerting Slack</label>
                                <input type="text" name="teamAlertingSlackUrl" id="teamAlertingSlackUrl" className="form-control" placeholder="API URL" value={this.getTeamConfigItem(row.Name, "AlertingSlackApi")} onChange={this.handleTeamInputChange.bind(this,row.Name,"AlertingSlackApi")} />
                            </div>
                            <div className="form-group">
                                <label htmlFor="inputNsApp" className="inputRg">Alerting Pagerduty</label>
                                <input type="text" name="teamAlertingPagerdutyUrl" id="teamAlertingPagerdutyUrl" className="form-control" placeholder="API URL" value={this.getTeamConfigItem(row.Name, "AlertingPagerdutyApi")} onChange={this.handleTeamInputChange.bind(this,row.Name,"AlertingPagerdutyApi")} />
                            </div>
                            <div className="toolbox">
                                <button type="submit" className="btn btn-primary bnt-k8s-namespace-create" disabled={this.stateUpdateButton()} onClick={this.updateTeamSettings.bind(this, row.Name)}>Save</button>
                            </div>
                        </form>
                    </div>
                )}


            </div>
        );
    }
}

export default Settings;
