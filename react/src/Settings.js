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

            userMessage: "",
            userError: "",

            teamMessage: {},
            teamError: {},

            settingConfig: {
                User: [],
                Team: []
            },

            user: {},

            team: {}
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

                if (jqxhr.Configuration) {
                    state.settingConfig = jqxhr.Configuration;
                }

                if (jqxhr.User) {
                    state.user = jqxhr.User;
                }

                if (jqxhr.Team) {
                    state.team = jqxhr.Team;
                }

                this.setState(state);
            }
        });
    }

    handlePersonalInputChange(name, event) {
        var state = this.state.user;
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

    updateUserSettings(e) {
        e.preventDefault();
        e.stopPropagation();

        this.setState({
            requestRunning: true,
            globalError: "",
            userMessage: "",
            userError: "",
            teamMessage: {},
            teamError: {}
        });

        let jqxhr = $.ajax({
            type: 'POST',
            url: "/api/settings/user",
            data: {
                config: this.state.user
            }
        }).done((jqxhr) => {
            this.setState({
                userMessage: "Personal settings updated",
                userError: ""
            });

        }).fail((jqxhr) => {
            var state = {
                userMessage: "",
                userError: ""
            };

            if (jqxhr.responseJSON.Message) {
                state.userError = jqxhr.responseJSON.Message;
            } else {
                state.userError = "Unknown error";
            }

            this.setState(state);
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
            userMessage: "",
            teamMessage: {},
            teamError: {}
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
                teamMessage: {},
                teamError: {},
            };
            state.teamMessage[team] = "Team " + team + " settings updated";
            this.setState(state);
        }).fail((jqxhr) => {
            console.log(jqxhr);

            var state = {
                teamMessage: {},
                teamError: {}
            };

            if (jqxhr.responseJSON.Message) {
                state.teamError[team] = jqxhr.responseJSON.Message;
            } else {
                state.teamError[team] = "Unknown error";
            }

            this.setState(state);
        }).always(() => {
            this.setState({
                requestRunning: false,
            });
        });

        this.handleXhr(jqxhr);
    }

    getUserConfigItem(name) {
        var ret = "";

        if (this.state.user && this.state.user[name]) {
            ret = this.state.user[name];
        }

        return ret;
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

    getTeamError(team) {
        if (this.state.teamError && this.state.teamError[team]) {
            return this.state.teamError[team];
        }
        return false
    }

    render() {
        if ((this.state.isStartupConfig || this.state.isStartupSettings) && this.state.globalError) {
            return (
                <div className="alert alert-danger">{this.state.globalError}</div>
            )
        }

        if (this.state.isStartupConfig || this.state.isStartupSettings) {
            return (
                <div>
                    <Spinner active={this.isStartup()}/>
                </div>
            );
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
                    <div className={this.state.userMessage === '' ? 'alert alert-success invisible' : 'alert alert-success'}>{this.state.userMessage}</div>
                    <div className={this.state.userError === '' ? 'alert alert-danger invisible' : 'alert alert-danger'}>{this.state.userError}</div>
                </div>
                <form method="post">
                    {this.state.settingConfig.User.map((setting, value) =>
                        <div className="form-group">
                            <label htmlFor="inputNsApp" className="inputRg">{setting.Label}</label>
                            <input type="text" name={setting.Name} id={setting.Name} className="form-control" placeholder={setting.Plaeholder} value={this.getUserConfigItem(setting.Name)} onChange={this.handlePersonalInputChange.bind(this, setting.Name)} />
                        </div>
                    )}
                    <div className="toolbox">
                        <button type="submit" className="btn btn-primary bnt-k8s-namespace-create" disabled={this.stateUpdateButton()} onClick={this.updateUserSettings.bind(this)}>Save</button>
                    </div>
                </form>


                {this.state.config.Teams.map((team, value) =>
                    <div>
                        <h2>Team {team.Name} settings</h2>
                        <div>
                            <div className={this.getTeamMessage(team.Name) === false ? 'alert alert-success invisible' : 'alert alert-success'}>{this.getTeamMessage(team.Name)}</div>
                            <div className={this.getTeamError(team.Name) === false ? 'alert alert-danger invisible' : 'alert alert-danger'}>{this.getTeamError(team.Name)}</div>
                        </div>
                        <form method="post">
                            {this.state.settingConfig.Team.map((setting, value) =>
                                <div className="form-group">
                                    <label htmlFor="inputNsApp" className="inputRg">{setting.Label}</label>
                                    <input type="text" name={setting.Name} id={setting.Name} className="form-control" placeholder={setting.Plaeholder} value={this.getTeamConfigItem(team.Name, setting.Name)} onChange={this.handleTeamInputChange.bind(this, team.Name, setting.Name)} />
                                </div>
                            )}
                            <div className="toolbox">
                                <button type="submit" className="btn btn-primary bnt-k8s-namespace-create" disabled={this.stateUpdateButton()} onClick={this.updateTeamSettings.bind(this, team.Name)}>Save</button>
                            </div>
                        </form>
                    </div>
                )}


            </div>
        );
    }
}

export default Settings;
