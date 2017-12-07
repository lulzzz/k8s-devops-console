import React, { Component } from 'react';
import $ from 'jquery';

class K8sNamespaceModalCreate extends Component {
    constructor(props) {
        super(props);

        this.state = {
            namespacePreview: "",
            nsEnvironment: "user",
            nsTeam: "",
            nsApp: "",
            buttonText: "Create namespace",
            buttonState: "disabled",
            message: ""
        };
    }

    createNamespace() {
        let oldButtonText = this.state.buttonText;
        this.setState({
            buttonState: "disabled",
            buttonText: "Saving...",
            message: ""
        });
        $.ajax({
            type: 'PUT',
            url: "/api/namespace",
            data: {
                nsEnvironment: this.state.nsEnvironment,
                nsAreaTeam: this.state.nsTeam,
                nsApp: this.state.nsApp
            }
        }).done((data) => {
            this.setState({
               nsApp: ""
            });

            if (this.props.callback) {
                this.props.callback()
            }
        }).fail((data) => {
            if (data.responseJSON && data.responseJSON.Message) {
                this.setState({
                    message: data.responseJSON.Message
                });
            }
        }).always(() => {
            this.setState({
                buttonState: "",
                buttonText: oldButtonText
            });
        });
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

    render() {
        if (this.state.nsTeam == '') {
            if (this.props.config.Teams.length > 0) {
                this.setState({nsTeam: this.props.config.Teams[0].Name});
            }
        }

        return (
            <div>
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
                                <form method="post">
                                    <div className={this.state.message === '' ? null : 'alert alert-danger'}>{this.state.message}</div>
                                    <div className="row">
                                        <div className="col-3">
                                            <label htmlFor="inputNsEnvironment">Environment</label>
                                            <select name="nsEnvironment" id="inputNsEnvironment" className="form-control" required value={this.state.nsEnvironment} onChange={this.handleNsEnvironmentChange.bind(this)}>
                                            {this.props.config.NamespaceEnvironments.map((row) =>
                                                <option key={row} value={row}>{row}</option>
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
                                                    {this.props.config.Teams.map((row) =>
                                                        <option key="team-{row.Name}" value={row.Name}>{row.Name}</option>
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
                                            <div className="p-3 mb-2 bg-light text-dark">
                                                <i>Preview: </i>
                                                {(() => {
                                                    switch (this.state.nsEnvironment) {
                                                        case "user":
                                                            return <span id="namespacePreview">user-{this.props.config.User.Username}-{this.state.nsApp}</span>;
                                                        case "team":
                                                            return <span id="namespacePreview">team-{this.state.nsTeam}-{this.state.nsApp}</span>;
                                                        default:
                                                            return <span id="namespacePreview">{this.state.nsEnvironment}-{this.state.nsApp}</span>;
                                                    }
                                                })()}
                                            </div>
                                        </div>
                                    </div>
                                </form>
                            </div>
                            <div className="modal-footer">
                                <button type="button" className="btn btn-secondary bnt-k8s-namespace-cancel" data-dismiss="modal">Cancel</button>
                                <button type="button" className="btn btn-primary bnt-k8s-namespace-create" disabled={this.state.buttonState} onClick={this.createNamespace.bind(this)}>{this.state.buttonText}</button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        );
    }
}

export default K8sNamespaceModalCreate;

