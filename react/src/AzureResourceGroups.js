import React from 'react';
import $ from 'jquery';
import onClickOutside from 'react-onclickoutside'

import BaseComponent from './BaseComponent';
import Spinner from './Spinner';

class K8sNamespace extends BaseComponent {
    constructor(props) {
        super(props);

        this.state = {
            globalMessage: "",
            globalError: "",
            searchValue: "",
            buttonText: "Create Azure ResourceGroup",
            requestRunning: false,

            resourceGroup: {
                team: "",
                name: "",
                location: "westeurope",
                personal: false,
                tag: {}
            },

            config: {
                User: {
                    Username: '',
                },
                Teams: [],
                NamespaceEnvironments: [],
                Quota: {},
                Azure: {
                    ResourceGroup: {
                        Tags: []
                    }
                }
            },
            isStartup: true
        };

        setInterval(() => {
            this.refresh()
        }, 10000);
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

                if (this.state.isStartup) {
                    this.setInputFocus();
                }

                this.setState({
                    config: jqxhr,
                    globalError: '',
                    isStartup: false
                });

                this.componentWillMount();
            }
        });
    }

    componentWillMount() {
        // select first team if no selection available
        if (this.state.resourceGroup.team === "") {
            if (this.state.config.Teams.length > 0) {
                let state = this.state;
                state.resourceGroup.team = this.state.config.Teams[0].Name
                this.setState(state);
            }
        }
    }

    componentDidMount() {
        this.loadConfig();
        this.setInputFocus();
    }

    refresh() {
        this.setState({
            globalMessage: ""
        });
    }

    createResourceGroup(e) {
        e.preventDefault();
        e.stopPropagation();

        let oldButtonText = this.state.buttonText;
        this.setState({
            requestRunning: true,
            buttonText: "Saving...",
            globalError: ""
        });

        let jqxhr = $.ajax({
            type: 'PUT',
            url: "/api/azure/resourcegroup",
            data: this.state.resourceGroup
        }).done((jqxhr) => {
            let state = this.state;
            state.globalMessage = "Azure ResourceGroup " + this.state.azResourceGroup + " created";
            state.resourceGroup.name = "";
            this.setState(state);
        }).always(() => {
            this.setState({
                requestRunning: false,
                buttonText: oldButtonText
            });
        });

        this.handleXhr(jqxhr);
    }

    stateCreateButton() {
        let state = "";

        if (this.state.requestRunning) {
            state = "disabled";
        } else {
            if (this.state.azResourceGroup === "" || this.state.azTeam === "" || this.state.azResourceGroupLocation === "") {
                state = "disabled"
            }
        }

        return state
    }

    handleResourceGroupInputChange(name, event) {
        var state = this.state;
        state.resourceGroup[name] = event.target.value;
        this.setState(state);
    }


    handleResourceGroupTagInputChange(name, event) {
        var state = this.state;
        state.resourceGroup["tag"][name] = event.target.value;
        this.setState(state);
    }

    getResourceGroupItem(name) {
        var ret = "";

        if (this.state.resourceGroup && this.state.resourceGroup[name]) {
            ret = this.state.resourceGroup[name];
        }

        return ret;
    }

    getResourceGroupTagItem(name) {
        var ret = "";

        if (this.state.resourceGroup.tag && this.state.resourceGroup.tag[name]) {
            ret = this.state.resourceGroup.tag[name];
        }

        return ret;
    }


    handleClickOutside() {
        this.setInputFocus();
    }

    azureResourceGroupTagConfig() {
        let ret = [];

        if (this.state.config.Azure.ResourceGroup.Tags) {
            ret = this.state.config.Azure.ResourceGroup.Tags
        }

        return ret;
    }

    render() {
        if (this.state.isStartup && this.state.globalError) {
            return (
                <div className="alert alert-danger">{this.state.globalError}</div>
            )
        }

        return (
            <div>
                <Spinner active={this.state.isStartup}/>
                <div>
                    <div className={this.state.globalError === '' ? null : 'alert alert-danger'}>{this.state.globalError}</div>
                    <div className={this.state.globalMessage === '' ? 'alert alert-success invisible' : 'alert alert-success'}>{this.state.globalMessage}</div>
                </div>

                <div>
                    <form method="post">
                        <div className="form-group">
                            <label htmlFor="inputNsAreaTeam">Team</label>
                            <select name="nsAreaTeam" id="inputNsAreaTeam" className="form-control namespace-area-team" value={this.getResourceGroupItem("team")} onChange={this.handleResourceGroupInputChange.bind(this, "team")}>
                                {this.state.config.Teams.map((row, value) =>
                                    <option key={row.Id} value={row.Name}>{row.Name}</option>
                                )}
                            </select>
                        </div>

                        <div className="form-group">
                            <label htmlFor="inputNsApp" className="inputRg">Azure ResourceGroup</label>
                            <input type="text" name="nsApp" id="inputRg" className="form-control" placeholder="ResourceGroup name" required value={this.getResourceGroupItem("name")} onChange={this.handleResourceGroupInputChange.bind(this, "name")} />
                        </div>
                        <div className="form-group">
                            <label htmlFor="inputNsApp" className="inputRgLocation">Azure Location</label>
                            <input type="text" name="nsApp" id="inputRgLocation" className="form-control" placeholder="ResourceGroup location" required value={this.getResourceGroupItem("location")} onChange={this.handleResourceGroupInputChange.bind(this, "location")} />
                        </div>

                        {this.azureResourceGroupTagConfig().map((setting, value) =>
                            <div className="form-group">
                                <label htmlFor="inputNsApp" className="inputRg">{setting.Label}</label>
                                <input type="text" name={setting.Name} id={setting.Name} className="form-control" placeholder={setting.Plaeholder} value={this.getResourceGroupTagItem(setting.Name)} onChange={this.handleResourceGroupTagInputChange.bind(this, setting.Name)} />
                            </div>
                        )}

                        <div className="form-group">
                            <div className="form-check">
                                <input type="checkbox" className="form-check-input" id="az-resourcegroup-personal" checked={this.getResourceGroupItem("personal")} onChange={this.handleResourceGroupInputChange.bind(this, "personal")} />
                                <label className="form-check-label" htmlFor="az-resourcegroup-personal">Personal ResourceGroup (only read access to team)</label>
                            </div>
                        </div>
                        <div className="toolbox">
                            <button type="submit" className="btn btn-primary bnt-k8s-namespace-create" disabled={this.stateCreateButton()} onClick={this.createResourceGroup.bind(this)}>{this.state.buttonText}</button>
                        </div>
                    </form>
                </div>

            </div>
        );
    }
}

export default onClickOutside(K8sNamespace);

