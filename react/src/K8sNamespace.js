import React from 'react';
import $ from 'jquery';

import BaseComponent from './BaseComponent';
import Spinner from './Spinner';
import K8sNamespaceModalDelete from './K8sNamespaceModalDelete';
import K8sNamespaceModalCreate from './K8sNamespaceModalCreate';

class K8sNamespace extends BaseComponent {
    constructor(props) {
        super(props);

        this.state = {
            isStartup: true,
            namespaces: [],
            confUser: {},
            config: {
                User: {
                    Username: '',
                },
                Teams: [],
                NamespaceEnvironments: []
            },
            selectedNamespace: [],
            selectedNamespaceDelete: [],
            namespacePreview: "",
            globalMessage: "",
            globalError: "",
            searchValue: "",
        };

        setInterval(() => {
            this.refresh()
        }, 10000);
    }

    loadNamespaces() {
        let jqxhr = $.get({
            url: '/api/namespace'
        }).done((jqxhr) => {
            this.setState({
                namespaces: jqxhr,
                globalError: '',
                isStartup: false
            });
        });

        this.handleXhr(jqxhr);
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
                    config: jqxhr
                });
            }
        });
    }

    componentDidMount() {
        this.loadConfig();
        this.loadNamespaces();
    }

    refresh() {
        this.loadNamespaces();
        this.setState({
            globalMessage: ""
        });
    }

    deleteNamespace(row) {
        this.setState({
            selectedNamespaceDelete: row
        });

        setTimeout(() => {
            $("#deleteQuestion").modal('show')
        }, 200);
    }

    createNamespace() {
        setTimeout(() => {
            $("#createQuestion").modal('show')
        }, 200);
    }

    selectNamespace(namespace) {
        this.setState({
            selectedNamespace: namespace
        });
    }

    renderRowOwner(row) {
        if (row.Name.match(/^user-[^-]+-.*/i)) {
            return <span><span className="badge badge-light">Personal Namespace</span></span>
        } else if (row.OwnerTeam !== "") {
            return <span><span className="badge badge-light">Team</span>{row.OwnerTeam}</span>
        } else if (row.OwnerUser !== "") {
            return <span><span className="badge badge-light">User</span>{row.OwnerUser}</span>
        }
    }

    handleNamespaceDeletion(namespace) {
        $("#deleteQuestion").modal('hide');
        this.refresh();
        this.setState({
            globalMessage: "Namespace \"" + namespace + "\" deleted"
        });
    }

    handleNamespaceCreation(namespace) {
        $("#createQuestion").modal('hide');
        this.refresh();
        this.setState({
            globalMessage: "Namespace \"" + namespace + "\" created"
        });
    }

    handleSearchChange(event) {
        this.setState({
            searchValue: event.target.value
        });
    }

    getNamespaces() {
        let ret = [];
        if (this.state.searchValue !== "") {
            let term =this.state.searchValue;
            ret = this.state.namespaces.filter((row) => {
                if (row.Name.includes(term)) {
                    return true;
                }

                if (row.OwnerTeam.includes(term)) {
                    return true;
                }

                if (row.OwnerUser.includes(term)) {
                    return true;
                }

                return false;
            });
        } else {
            ret = this.state.namespaces;
        }

        ret = ret.sort(function(a,b) {
            return a.Name >= b.Name;
        });

        return ret;
    }

    render() {
        if (this.state.globalError) {
            return (
                <div className="alert alert-danger">{this.state.globalError}</div>
            )
        }

        let self = this;
        let namespaces = this.getNamespaces();
        return (
            <div>
                <Spinner active={this.state.isStartup}/>
                <div className="container-toolbar-main">
                    <div className={this.state.globalMessage === '' ? 'alert alert-success invisible' : 'alert alert-success'}>{this.state.globalMessage}</div>
                    <input type="text" className="form-control search-input" placeholder="Search" value={this.state.searchValue} onChange={this.handleSearchChange.bind(this)} />
                    <div className="clearfix"></div>
                </div>
                <table className="table table-hover table-sm">
                    <colgroup>
                        <col width="*" />
                        <col width="200rem" />
                        <col width="200rem" />
                        <col width="100rem" />
                        <col width="80rem" />
                    </colgroup>
                    <thead>
                    <tr>
                        <th>Namespace</th>
                        <th>Owner</th>
                        <th>Created</th>
                        <th>Status</th>
                        <th className="toolbox">
                            <button type="button" className="btn btn-primary" onClick={this.createNamespace.bind(this)}>Create</button>
                        </th>
                    </tr>
                    </thead>
                    <tfoot>
                    <tr>
                        <td className="toolbox" colSpan="5">
                            <button type="button" className="btn btn-primary" onClick={this.createNamespace.bind(this)}>Create</button>
                        </td>
                    </tr>
                    </tfoot>
                    <tbody>
                    {namespaces.map((row) =>
                        <tr key={row.Name} onClick={this.selectNamespace.bind(this, row)}>
                            <td>{row.Name}</td>
                            <td>
                                {this.renderRowOwner(row)}
                            </td>
                            <td><div title={row.Created}>{row.CreatedAgo}</div></td>
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
                            <td className="toolbox">
                                {(() => {
                                    if (row.Deleteable) {
                                        switch (row.Status) {
                                            case "Terminating":
                                                return <button type="button" className="btn btn-danger"
                                                               disabled>Delete</button>;
                                            default:
                                                return <button type="button" className="btn btn-danger" onClick={self.deleteNamespace.bind(self, row)}>Delete</button>;

                                        }
                                    }
                                })()}

                            </td>
                        </tr>
                    )}
                    </tbody>
                </table>

                <K8sNamespaceModalDelete config={this.state.config} namespace={this.state.selectedNamespaceDelete} callback={this.handleNamespaceDeletion.bind(this)} />
                <K8sNamespaceModalCreate config={this.state.config} callback={this.handleNamespaceCreation.bind(this)} />
            </div>
        );
    }
}

export default K8sNamespace;

