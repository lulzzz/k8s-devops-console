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
            buttonState: "",
            azTeam: "",
            azResourceGroup: "",
            azResourceGroupLocation: "westeurope",
            azResourceGroupPersonal: false,
            config: {
                User: {
                    Username: '',
                },
                Teams: [],
                NamespaceEnvironments: [],
                Quota: {}
            },
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
        if (this.state.azTeam === "") {
            if (this.state.config.Teams.length > 0) {
                this.setState({azTeam: this.state.config.Teams[0].Name});
            }
        }
    }

    componentDidMount() {
        this.loadConfig();
    }

    refresh() {
        this.setState({
            globalMessage: ""
        });
    }

    handleAzTeamChange(event) {
        this.setState({
            azTeam: event.target.value
        });
    }

    handleAzResourceGroup(event) {
        this.setState({
            azResourceGroup: event.target.value
        });
    }

    handleAzResourceGroupLocation(event) {
        this.setState({
            azResourceGroupLocation: event.target.value
        });
    }

    handleAzResourceGroupPersonal(event) {
        this.setState({
            azResourceGroupPersonal: event.target.checked
        });
    }

    createResourceGroup() {
        let oldButtonText = this.state.buttonText;
        this.setState({
            buttonState: "disabled",
            buttonText: "Saving...",
            globalError: ""
        });

        let jqxhr = $.ajax({
            type: 'PUT',
            url: "/api/azure/resourcegroup",
            data: {
                team: this.state.azTeam,
                resourceGroupName: this.state.azResourceGroup,
                location: this.state.azResourceGroupLocation,
                personal: this.state.azResourceGroupPersonal
            }
        }).done((jqxhr) => {
            this.setState({
                globalMessage: "Azure ResourceGroup " + this.state.azResourceGroup + " created",
                azResourceGroup: ""
            });
        }).always(() => {
            this.setState({
                buttonState: "",
                buttonText: oldButtonText
            });
        });

        this.handleXhr(jqxhr);
    }

    getResourceGroups() {
        return [];
    }

    handleClickOutside() {

    }

    render() {
        if (this.state.isStartup && this.state.globalError) {
            return (
                <div className="alert alert-danger">{this.state.globalError}</div>
            )
        }

        if (this.state.isStartup) {
            return (
                <div>
                    <Spinner active={this.state.isStartup}/>
                </div>
            )
        }

        return (
            <div>
                <div>
                    <div className={this.state.globalError === '' ? null : 'alert alert-danger'}>{this.state.globalError}</div>
                    <div className={this.state.globalMessage === '' ? 'alert alert-success invisible' : 'alert alert-success'}>{this.state.globalMessage}</div>
                </div>

                <div>
                    <form method="post">
                        <div className="form-group">
                            <label htmlFor="inputNsAreaTeam">Team</label>
                            <select name="nsAreaTeam" id="inputNsAreaTeam" className="form-control namespace-area-team" value={this.state.azTeam} onChange={this.handleAzTeamChange.bind(this)}>
                                {this.state.config.Teams.map((row, value) =>
                                    <option key={row.Id} value={row.Name}>{row.Name}</option>
                                )}
                            </select>
                        </div>

                        <div className="form-group">
                            <label htmlFor="inputNsApp" className="inputNsApp">Azure ResourceGroup</label>
                            <input type="text" name="nsApp" id="inputNsApp" className="form-control" placeholder="Name" required value={this.state.azResourceGroup} onChange={this.handleAzResourceGroup.bind(this)} />
                        </div>
                        <div className="form-group">
                            <label htmlFor="inputNsApp" className="inputNsApp">Azure Location</label>
                            <input type="text" name="nsApp" id="inputNsApp" className="form-control" placeholder="Name" required value={this.state.azResourceGroupLocation} onChange={this.handleAzResourceGroupLocation.bind(this)} />
                        </div>
                        <div className="form-group">
                            <div className="form-check">
                                <input type="checkbox" className="form-check-input" id="az-resourcegroup-personal" checked={this.state.azResourceGroupPersonal} onChange={this.handleAzResourceGroupPersonal.bind(this)} />
                                <label className="form-check-label" htmlFor="az-resourcegroup-personal">Personal namespace (only read access to team)</label>
                            </div>
                        </div>
                    </form>

                    <div className="toolbox">
                        <button type="button" className="btn btn-primary bnt-k8s-namespace-create" disabled={this.state.buttonState} onClick={this.createResourceGroup.bind(this)}>{this.state.buttonText}</button>
                    </div>

                </div>

            </div>
        );
    }
}

export default onClickOutside(K8sNamespace);

