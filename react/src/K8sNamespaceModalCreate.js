import React from 'react';
import $ from 'jquery';
import {CopyToClipboard} from 'react-copy-to-clipboard';

import BaseComponent from './BaseComponent';

class K8sNamespaceModalCreate extends BaseComponent {
    constructor(props) {
        super(props);

        this.state = {
            namespacePreview: "",
            buttonText: "Create namespace",
            buttonState: "disabled",
            globalError: "",

            namespace: {
                environment: "",
                app: "",
                team: "",
                description: "",
                label: {}
            }
        };
    }

    createNamespace(e) {
        e.preventDefault();
        e.stopPropagation();

        let oldButtonText = this.state.buttonText;
        this.setState({
            buttonState: "disabled",
            buttonText: "Saving...",
            globalError: ""
        });

        let jqxhr = $.ajax({
            type: 'PUT',
            url: "/api/namespace",
            data: this.state.namespace
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

    handleNamespaceInputChange(name, event) {
        var state = this.state;
        state.namespace[name] = event.target.value;
        this.setState(state);

        this.handleButtonState()
    }


    handleNamespaceLabelInputChange(name, event) {
        var state = this.state;
        state.namespace["label"][name] = event.target.value;
        this.setState(state);
    }

    getNamespaceItem(name) {
        var ret = "";

        if (this.state.namespace && this.state.namespace[name]) {
            ret = this.state.namespace[name];
        }

        return ret;
    }

    getNamespaceLabelItem(name) {
        var ret = "";

        if (this.state.namespace.label && this.state.namespace.label[name]) {
            ret = this.state.namespace.label[name];
        }

        return ret;
    }

    handleButtonState(event) {
        let buttonState = "disabled";

        if (this.state.namespace.environment != "" && this.state.namespace.app != "" && this.state.namespace.team != "") {
            buttonState = ""
        }

        this.setState({
            buttonState: buttonState,
        });
    }

    handleNsDescriptionChange(event) {
        let state = this.state;
        state.namespace.description = event.target.value;
        this.setState(state);
    }

    previewNamespace() {
        let namespace = "";

        switch (this.state.namespace.environment) {
            case "user":
                namespace = "user-" + this.props.config.User.Username + "-" + this.state.namespace.app;
                break;
            case "team":
                namespace = "team-" + this.state.namespace.team + "-" + this.state.namespace.app;
                break;
            default:
                namespace = this.state.namespace.environment + "-" + this.state.namespace.app;
                break;
        }

        return namespace.toLowerCase().replace(/_/g, "");
    }

    componentWillMount() {
        let state = this.state;

        // select first team if no selection available
        if (this.state.namespace.team === "") {
            if (this.props.config.Teams.length > 0) {
                state.namespace.team = this.props.config.Teams[0].Name;
            }
        }

        if (this.state.namespace.environment === "") {
            if (this.props.config.NamespaceEnvironments.length > 0) {
                console.log(this.props.config.NamespaceEnvironments);
                state.namespace.environment = this.props.config.NamespaceEnvironments[0].Environment;
            }
        }

        this.setState(state);
    }


    kubernetesLabelConfig() {
        let ret = [];

        if (this.props.config.Kubernetes.Namespace.Labels) {
            ret = this.props.config.Kubernetes.Namespace.Labels;
        }

        return ret;
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
                                            <select name="nsEnvironment" id="inputNsEnvironment" className="form-control" required value={this.getNamespaceItem("environment")} onChange={this.handleNamespaceInputChange.bind(this, "environment")}>
                                            {this.props.config.NamespaceEnvironments.map((row) =>
                                                <option key={row.Environment} value={row.Environment}>{row.Environment} ({row.Description})</option>
                                            )}
                                            </select>
                                        </div>
                                        <div>
                                            <label htmlFor="inputNsAreaTeam">Team</label>
                                            <select name="nsAreaTeam" id="inputNsAreaTeam" className="form-control namespace-area-team" value={this.getNamespaceItem("team")} onChange={this.handleNamespaceInputChange.bind(this, "team")}>
                                                {this.props.config.Teams.map((row, value) =>
                                                    <option key={row.Id} value={row.Name}>{row.Name}</option>
                                                )}
                                            </select>
                                        </div>
                                        <div className="col">
                                            <label htmlFor="inputNsApp" className="inputNsApp">Application</label>
                                            <input type="text" name="nsApp" id="inputNsApp" className="form-control" placeholder="Name" required value={this.getNamespaceItem("app")} onChange={this.handleNamespaceInputChange.bind(this, "app")} />
                                        </div>
                                    </div>
                                    <div className="row">
                                        <div className="col">
                                            <input type="text" name="nsDescription" id="inputNsDescription" className="form-control" placeholder="Description" value={this.getNamespaceItem("description")} onChange={this.handleNamespaceInputChange.bind(this, "description")} />
                                        </div>
                                    </div>

                                    {this.kubernetesLabelConfig().map((setting, value) =>
                                        <div className="form-group">
                                            <label htmlFor="inputNsApp" className="inputRg">{setting.Label}</label>
                                            <input type="text" name={setting.Name} id={setting.Name} className="form-control" placeholder={setting.Plaeholder} value={this.getNamespaceLabelItem(setting.Name)} onChange={this.handleNamespaceLabelInputChange.bind(this, setting.Name)} />
                                        </div>
                                    )}


                                    <div className="row">
                                        <div className="col">
                                            <div className="p-3 mb-2 bg-light text-dark">
                                                <i>Preview: </i><span id="namespacePreview">{this.previewNamespace()}</span>
                                                <CopyToClipboard text={this.previewNamespace()}>
                                                    <button className="button-copy" onClick={this.handlePreventEvent.bind(this)}></button>
                                                </CopyToClipboard>
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

