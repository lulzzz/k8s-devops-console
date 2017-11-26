import React, { Component } from 'react';
import $ from 'jquery';

class K8sNamespace extends Component {
    constructor(props) {
        super(props);

        this.state = {
            namespaces: [],
            confUser: {},
            config: {
                User: {},
                Teams: [],
                NamespaceEnvironments: []
            },
            selectedNamespace: [],
            namespacePreview: "",
            globalMessage: "",
            deleteButtonState: "",
            deleteButtonText: "Delete namespace",
            createEnvironment: "user",
            createUser: "",
            createTeam: "",
            createApp: "",
            createButtonText: "Create namespace",
            createButtonState: "disabled",
            createMessage: ""
        };

        setInterval(() => {
            this.refresh()
        }, 10000);
    }

    loadNamespaces() {
        $.get({
            url: '/api/namespace'
        }).done((data) => {

            data = data.sort(function(a,b) {
                return a.Name >= b.Name;
            });

            this.setState({
                namespaces: data
            });
        });
    }

    loadConfig() {
        $.get({
            url: '/api/_app/config'
        }).done((data) => {
            let username = "";
            let team = "";

            if (data) {

                if (data.User && data.User.Username) {
                    username = data.User.Username;
                }

                if (data.Teams && data.Teams.length) {
                    team = data.Teams[0].Name
                }

                this.setState({
                    config: data,
                    createUser: username,
                    createTeam: team
                });
            }

        });
    }

    componentDidMount() {
        this.loadNamespaces();
        this.loadConfig();
    }

    componentWillMount(){
        window.k8sNamespaces = (data) => {
            this.loadNamespaces()
        };
    }

    refresh() {
        this.loadNamespaces();
    }

    deleteNamespace(row) {
        this.setState({
           selectedNamespace: row
        });

        setTimeout(() => {
            $("#deleteQuestion").modal('show')
        }, 200);
    }

    doDeleteNamespace() {
        if (!this.state.selectedNamespace) {
            return
        }

        let oldButtonText = this.state.createButtonText;
        this.setState({
            deleteButtonState: "disabled",
            deleteButtonText: "Deleting..."
        });

        $.ajax({
            type: 'DELETE',
            url: "/api/namespace/?" + $.param({"namespace": this.state.selectedNamespace.Name})
        }).done(() => {
            $("#deleteQuestion").modal('hide');
            this.setState({
                globalMessage: "Namespace \"" + this.state.selectedNamespace.Name + "\" deleted"
            });
            this.refresh();
        }).always(() => {
            this.setState({
                deleteButtonState: "",
                deleteButtonText: oldButtonText
            });
        })
    }

    createNamespace() {
        setTimeout(() => {
            $("#createQuestion").modal('show')
        }, 200);
    }

    doCreateNamespace() {
        let oldButtonText = this.state.deleteButtonText;
        this.setState({
            createButtonState: "disabled",
            createButtonText: "Saving...",
            createMessage: ""
        });
        $.ajax({
            type: 'PUT',
            url: "/api/namespace",
            data: {
                nsEnvironment: this.state.createEnvironment,
                nsAreaTeam: this.state.createTeam,
                nsApp: this.state.createApp
            }
        }).done((data) => {
            $("#createQuestion").modal('hide');

            this.setState({
                createApp: "",
            });

            if (data && data.Namespace) {
                this.setState({
                    globalMessage: "Namespace \"" + data.Namespace + "\" created"
                })
            }
            this.loadNamespaces()
        }).fail((data) => {
            if (data.responseJSON && data.responseJSON.Message) {
                this.setState({
                    createMessage: data.responseJSON.Message
                });
            }
        }).always(() => {
            this.setState({
                createButtonState: "",
                createButtonText: oldButtonText
            });
        });
    }

    handleCreateNsEnvironmentChange(event) {
        this.setState({
            createEnvironment: event.target.value
        });
    }

    handleCreateNsTeamChange(event) {
        this.setState({
            createTeam: event.target.value
        });
    }

    handleCreateNsUserChange(event) {
    }

    handleCreateNsAppChange(event) {
        let buttonState = "disabled";

        if (event.target.value) {
            buttonState = "";
        }
        
        this.setState({
            createApp: event.target.value,
            createButtonState: buttonState,
        });
    }

    renderRowOwner(row) {
        if (row.OwnerTeam !== "") {
            return <span><span className="badge badge-light">Team</span> {row.OwnerTeam}</span>
        } else if (row.OwnerUser !== "") {
            return <span><span className="badge badge-light">User</span> {row.OwnerUser}</span>
        }
    }

    render() {
        let self = this;
        if (this.state.namespaces) {
            return (
                <div>
                    <div className="container-toolbar">
                        <button type="button" className="btn btn-primary bnt-ns-create" onClick={this.createNamespace.bind(this)}>Create</button>
                    </div>
                    <div className={this.state.globalMessage === '' ? null : 'alert alert-success'}>{this.state.globalMessage}</div>
                    <table className="table table-hover table-sm">
                        <thead>
                        <tr>
                            <th>Namespace</th>
                            <th>Owner</th>
                            <th>Status</th>
                            <th>Created</th>
                            <th className="toolbox"></th>
                        </tr>
                        </thead>
                        <tbody>
                        {this.state.namespaces.map((row) =>
                            <tr key={row.Name}>
                                <td>{row.Name}</td>
                                <td>
                                    {this.renderRowOwner(row)}
                                </td>
                                <td>
                                    {(() => {
                                        switch (row.Status) {
                                            case "Terminating":
                                                return <span className="badge badge-danger">{row.Status}</span>;
                                            case "Active":
                                                return <span className="badge badge-success">{row.Status}</span>;
                                            default:
                                                return <span className="badge badge-warning">{row.Status}</span>;
                                        }
                                    })()}
                                </td>
                                <td><div title={row.Created}>{row.CreatedAgo}</div></td>
                                <td className="toolbox">
                                    {(() => {
                                        switch (row.Status) {
                                            case "Terminating":
                                                return <button type="button" className="btn btn-danger" disabled>Delete</button>;
                                            default:
                                                return <button type="button" className="btn btn-danger" onClick={self.deleteNamespace.bind(self, row)}>Delete</button>;
                                        }
                                    })()}

                                </td>
                            </tr>
                        )}
                        </tbody>
                    </table>

                    <div className="modal fade" id="deleteQuestion" tabIndex="-1" role="dialog" aria-labelledby="deleteQuestion">
                        <div className="modal-dialog" role="document">
                            <div className="modal-content">
                                <div className="modal-header">
                                    <h5 className="modal-title" id="exampleModalLabel">Delete namespace</h5>
                                    <button type="button" className="close" data-dismiss="modal" aria-label="Close">
                                        <span aria-hidden="true">&times;</span>
                                    </button>
                                </div>
                                <div className="modal-body">
                                    Do you really want to delete namespace <strong className="k8s-namespace">{this.state.selectedNamespace.Name}</strong>?
                                </div>
                                <div className="modal-footer">
                                    <button type="button" className="btn btn-primary bnt-k8s-namespace-cancel" data-dismiss="modal">Cancel</button>
                                    <button type="button" className="btn btn-secondary bnt-k8s-namespace-delete" disabled={this.state.deleteButtonState} onClick={this.doDeleteNamespace.bind(this)}>{this.state.deleteButtonText}</button>
                                </div>
                            </div>
                        </div>
                    </div>

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
                                        <div className={this.state.createMessage === '' ? null : 'alert alert-danger'}>{this.state.createMessage}</div>
                                        <div className="row">
                                            <div className="col-3">
                                                <label htmlFor="inputNsEnvironment">Environment</label>
                                                <select name="nsEnvironment" id="inputNsEnvironment" className="form-control" required value={this.state.createEnvironment} onChange={this.handleCreateNsEnvironmentChange.bind(this)}>
                                                {this.state.config.NamespaceEnvironments.map((row) =>
                                                    <option key={row} value={row}>{row}</option>
                                                )}
                                                </select>
                                            </div>
                                            <div>
                                                <div className={this.state.createEnvironment === 'user' ? null : 'hidden'}>
                                                    <label htmlFor="inputNsAreaUser">User</label>
                                                    <input type="text" name="nsAreaUser" id="inputNsAreaUser" className="form-control namespace-area-user" value={this.state.createUser} onChange={this.handleCreateNsUserChange.bind(this)} disabled />
                                                </div>
                                                <div className={this.state.createEnvironment === 'user' ? 'hidden' : null}>
                                                    <label htmlFor="inputNsAreaTeam">Team</label>
                                                    <select name="nsAreaTeam" id="inputNsAreaTeam" className="form-control namespace-area-team" value={this.state.createTeam} onChange={this.handleCreateNsTeamChange.bind(this)}>
                                                        {this.state.config.Teams.map((row) =>
                                                            <option key="team-{row.Name}" value={row.Name}>{row.Name}</option>
                                                        )}
                                                    </select>
                                                </div>
                                            </div>
                                            <div>
                                            </div>
                                            <div className="col">
                                                <label htmlFor="inputNsApp" className="inputNsApp">Application</label>
                                                <input type="text" name="nsApp" id="inputNsApp" className="form-control" placeholder="Name" required value={this.state.createApp} onChange={this.handleCreateNsAppChange.bind(this)} />
                                            </div>
                                        </div>
                                        <div className="row">
                                            <div className="col">
                                                <div className="p-3 mb-2 bg-light text-dark">
                                                    <i>Preview: </i>
                                                    {(() => {
                                                        switch (this.state.createEnvironment) {
                                                            case "user":
                                                                return <span id="namespacePreview">user-{this.state.createUser}-{this.state.createApp}</span>;
                                                            case "team":
                                                                return <span id="namespacePreview">team-{this.state.createUser}-{this.state.createApp}</span>;
                                                            default:
                                                                return <span id="namespacePreview">{this.state.createEnvironment}-{this.state.createApp}</span>;
                                                        }
                                                    })()}
                                                </div>
                                            </div>
                                        </div>
                                    </form>
                                </div>
                                <div className="modal-footer">
                                    <button type="button" className="btn btn-secondary bnt-k8s-namespace-cancel" data-dismiss="modal">Cancel</button>
                                    <button type="button" className="btn btn-primary bnt-k8s-namespace-create" disabled={this.state.createButtonState} onClick={this.doCreateNamespace.bind(this)}>{this.state.createButtonText}</button>
                                </div>
                            </div>
                        </div>
                    </div>

                </div>
            );
        } else {
            return (
                <div>
                    <div className="container-toolbar">
                        <button type="button" className="btn btn-primary bnt-ns-create" onClick={this.refresh.bind(this)}>Create</button>
                    </div>
                    <div className="alert alert-info">No namespaces found</div>
                </div>
            )
        }
    }
}

export default K8sNamespace;

