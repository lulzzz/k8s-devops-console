import React from 'react';
import $ from 'jquery';

import BaseComponent from './BaseComponent';

class K8sNamespaceModalCreate extends BaseComponent {
    constructor(props) {
        super(props);

        this.state = {
            namespacePreview: "",
            nsEnvironment: "user",
            nsTeam: "",
            nsApp: "",
            nsDescription: "",
            buttonText: "Create namespace",
            buttonState: "disabled",
            globalError: ""
        };
    }

    createNamespace() {
        let oldButtonText = this.state.buttonText;
        this.setState({
            buttonState: "disabled",
            buttonText: "Saving...",
            globalError: ""
        });

        let jqxhr = $.ajax({
            type: 'PUT',
            url: "/api/namespace",
            data: {
                nsEnvironment: this.state.nsEnvironment,
                nsAreaTeam: this.state.nsTeam,
                nsApp: this.state.nsApp,
                description: this.state.nsDescription
            }
        }).done((jqxhr) => {
            this.setState({
               nsApp: "",
                nsDescription: ""
            });

            let namespace = ""
            if (jqxhr.Namespace) {
                namespace = jqxhr.Namespace;
            }

            if (this.props.callback) {
                this.props.callback(namespace)
            }
        }).always(() => {
            this.setState({
                buttonState: "",
                buttonText: oldButtonText
            });
        });

        this.handleXhr(jqxhr);
    }

    handleNsEnvironmentChange(event) {
        this.setState({
            nsEnvironment: event.target.value
        });
    }

    handleNsTeamChange(event) {
        this.setState({
            nsTeam: event.target.value
        });
    }

    handleNsUserChange(event) {
    }

    handleNsAppChange(event) {
        let buttonState = "disabled";

        if (event.target.value) {
            buttonState = "";
        }
        
        this.setState({
            nsApp: event.target.value,
            buttonState: buttonState,
        });
    }

    handleNsDescriptionChange(event) {
        this.setState({
            nsDescription: event.target.value
        });
    }

    previewNamespace() {
        let namespace = "";

        switch (this.state.nsEnvironment) {
            case "user":
                namespace = "user-" + this.props.config.User.Username + "-" + this.state.nsApp;
                break;
            case "team":
                namespace = "team-" + this.state.nsTeam + "-" + this.state.nsApp;
                break;
            default:
                namespace = this.state.nsEnvironment + "-" + this.state.nsApp;
                break;
        }

        namespace = namespace.toLowerCase().replace(/_/g, "");

        return <span id="namespacePreview">{namespace}</span>;
    }

    componentWillMount() {
        // select first team if no selection available
        if (this.state.nsTeam === "") {
            if (this.props.config.Teams.length > 0) {
                this.setState({nsTeam: this.props.config.Teams[0].Name});
            }
        }
    }

    render() {
        return (
            <div>
                <form method="post">
                <div className="modal fade" id="createQuestion" tabIndex="-1" role="dialog" aria-labelledby="createQuestion" aria-hidden="true">
                    <div className="modal-dialog" role="document">
                        <div className="modal-content">
                            <div className="modal-header">
                                <h5 className="modal-title" id="exampleModalLabel">Create namespace</h5>
                                <button type="button" className="close" data-dismiss="modal" aria-label="Close">
                                    <span aria-hidden="true">&times;</span>
                                </button>
                            </div>
                                <div className="modal-body">
                                    <div className={this.state.globalError === '' ? null : 'alert alert-danger'}>{this.state.globalError}</div>
                                    <div className="row">
                                        <div className="col-3">
                                            <label htmlFor="inputNsEnvironment">Environment</label>
                                            <select name="nsEnvironment" id="inputNsEnvironment" className="form-control" required value={this.state.nsEnvironment} onChange={this.handleNsEnvironmentChange.bind(this)}>
                                            {this.props.config.NamespaceEnvironments.map((row) =>
                                                <option key={row.Environment} value={row.Environment}>{row.Environment} ({row.Description})</option>
                                            )}
                                            </select>
                                        </div>
                                        <div>
                                            <div className={this.state.nsEnvironment === 'user' ? null : 'hidden'}>
                                                <label htmlFor="inputNsAreaUser">User</label>
                                                <input type="text" name="nsAreaUser" id="inputNsAreaUser" className="form-control namespace-area-user" value={this.props.config.User.Username} onChange={this.handleNsUserChange.bind(this)} disabled />
                                            </div>
                                            <div className={this.state.nsEnvironment === 'user' ? 'hidden' : null}>
                                                <label htmlFor="inputNsAreaTeam">Team</label>
                                                <select name="nsAreaTeam" id="inputNsAreaTeam" className="form-control namespace-area-team" value={this.state.nsTeam} onChange={this.handleNsTeamChange.bind(this)}>
                                                    {this.props.config.Teams.map((row, value) =>
                                                        <option key={row.Id} value={row.Name}>{row.Name}</option>
                                                    )}
                                                </select>
                                            </div>
                                        </div>
                                        <div className="col">
                                            <label htmlFor="inputNsApp" className="inputNsApp">Application</label>
                                            <input type="text" name="nsApp" id="inputNsApp" className="form-control" placeholder="Name" required value={this.state.nsApp} onChange={this.handleNsAppChange.bind(this)} />
                                        </div>
                                    </div>
                                    <div className="row">
                                        <div className="col">
                                            <input type="text" name="nsDescription" id="inputNsDescription" className="form-control" placeholder="Description" value={this.state.nsDescription} onChange={this.handleNsDescriptionChange.bind(this)} />
                                        </div>
                                    </div>
                                    <div className="row">
                                        <div className="col">
                                            <div className="p-3 mb-2 bg-light text-dark">
                                                <i>Preview: </i>{this.previewNamespace()}
                                            </div>
                                        </div>
                                    </div>
                                </div>
                                <div className="modal-footer">
                                    <button type="button" className="btn btn-secondary bnt-k8s-namespace-cancel" data-dismiss="modal">Cancel</button>
                                    <button type="submit" className="btn btn-primary bnt-k8s-namespace-create" disabled={this.state.buttonState} onClick={this.createNamespace.bind(this)}>{this.state.buttonText}</button>
                                </div>
                            </div>
                        </div>
                    </div>
                </form>
            </div>
        );
    }
}

export default K8sNamespaceModalCreate;

