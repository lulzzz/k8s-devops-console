import React from 'react';
import $ from 'jquery';
import onClickOutside from 'react-onclickoutside'

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
                NamespaceEnvironments: [],
                Quota: {}
            },
            namespaceDescriptionEdit: false,
            namespaceDescriptionEditValue: "",
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

    handleClickOutside() {
        this.handleDescriptionEditClose();
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

    handleNamespaceClick(row, event) {
        // close descripton if clicked somewhere else
        if (this.state.namespaceDescriptionEdit !== false && this.state.namespaceDescriptionEdit !== row.Name) {
            this.handleDescriptionEditClose();
        }

        this.setState({
            selectedNamespace: row
        });
        event.stopPropagation();
    }

    resetPermissions(namespace) {
        let jqxhr = $.ajax({
            type: 'POST',
            url: "/api/mgmt/namespace/resetpermissions/" + encodeURI(namespace.Name)
        }).done((jqxhr) => {
            if (jqxhr.Message) {
                this.setState({
                    globalMessage: jqxhr.Message
                });
            }
        });

        this.handleXhr(jqxhr);
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

    handleDescriptionEditClose() {
        this.setState({
            namespaceDescriptionEdit: false,
            namespaceDescriptionEditValue: ""
        });
    }

    handleDescriptionEdit(row, event) {
        this.setState({
            namespaceDescriptionEdit: row.Name,
            namespaceDescriptionEditValue: row.Description
        });

        setTimeout(() => {
            $(".description-edit").focus();
            $(".description-edit").each((index, el) => {
                // place cursor at end
                el.selectionStart =  el.selectionEnd = 10000;
            });
        }, 150)
    }

    handleDescriptionChange(event) {
        this.setState({
            namespaceDescriptionEditValue: event.target.value
        });
    }

    handleDescriptionSubmit(event) {
        let jqxhr = $.ajax({
            type: 'POST',
            url: "/api/mgmt/namespace/description/" + encodeURI(this.state.namespaceDescriptionEdit),
            data: {
                description: this.state.namespaceDescriptionEditValue
            }
        }).done((jqxhr) => {
            if (jqxhr.Message) {
                this.setState({
                    globalMessage: jqxhr.Message,
                    namespaceDescriptionEdit: false
                });
                this.refresh();
            }
        });

        this.handleXhr(jqxhr);
        event.preventDefault();
        event.stopPropagation();
        return false;
    }

    getNamespaces() {
        let ret = [];
        if (this.state.searchValue !== "") {
            let term = this.state.searchValue.replace(/[.?*+^$[\]\\(){}|-]/g, "\\$&");
            let re = new RegExp(term, "i");

            ret = this.state.namespaces.filter((row) => {
                if (row.Name.search(re) !== -1) return true;
                if (row.OwnerTeam.search(re) !== -1) return true;
                if (row.OwnerUser.search(re) !== -1) return true;
                if (row.Description.search(re) !== -1) return true;

                return false;
            });
        } else {
            ret = this.state.namespaces;
        }

        ret = ret.sort(function(a,b) {
            if(a.Name < b.Name) return -1;
            if(a.Name > b.Name) return 1;
            return 0;
        });

        return ret;
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

        let self = this;
        let namespaces = this.getNamespaces();
        return (
            <div onClick={this.handleDescriptionEditClose.bind(this)}>
                <div className="container-toolbar-main">
                    <div className="floating-message">
                        <div className={this.state.globalError === '' ? null : 'alert alert-danger'}>{this.state.globalError}</div>
                        <div className={this.state.globalMessage === '' ? 'alert alert-success invisible' : 'alert alert-success'}>{this.state.globalMessage}</div>
                    </div>
                    <input type="text" className="form-control search-input" placeholder="Search" value={this.state.searchValue} onChange={this.handleSearchChange.bind(this)}/>
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
                        <td colSpan="3">
                            <small>Namespace quota: {this.state.config.Quota.team === 0 ? 'unlimited' : this.state.config.Quota.team} team / {this.state.config.Quota.user === 0 ? 'unlimited' : this.state.config.Quota.user} personal</small>
                        </td>
                        <td className="toolbox" colSpan="3">
                            <button type="button" className="btn btn-primary" onClick={this.createNamespace.bind(this)}>Create</button>
                        </td>
                    </tr>
                    </tfoot>
                    <tbody>
                    {namespaces.map((row) =>
                        <tr key={row.Name} className="k8s-namespace" onClick={this.handleNamespaceClick.bind(this, row)}>
                            <td>
                                {row.Name}<br/>
                                {(() => {
                                   if (this.state.namespaceDescriptionEdit === row.Name) {
                                       return <form onSubmit={this.handleDescriptionSubmit.bind(this)}>
                                           <input type="text" className="form-control description-edit" placeholder="Description" value={this.state.namespaceDescriptionEditValue} onChange={this.handleDescriptionChange.bind(this)}/>
                                       </form>
                                   } else {
                                       return <small className="form-text text-muted editable description" onClick={this.handleDescriptionEdit.bind(this, row)}>{row.Description ? row.Description : <i>no description set</i>}</small>
                                   }
                                })()}
                            </td>
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
                                <br/>
                                <span className={row.Deleteable ? 'hidden' : 'badge badge-info'}>Not deletable</span>
                            </td>
                            <td className="toolbox">
                                {(() => {
                                    switch (row.Status) {
                                    case "Terminating":
                                        return <div></div>
                                    default:
                                        return (
                                            <div className="btn-group" role="group">
                                                <button id="btnGroupDrop1" type="button"
                                                        className="btn btn-secondary dropdown-toggle"
                                                        data-toggle="dropdown" aria-haspopup="true"
                                                        aria-expanded="false">
                                                    Action
                                                </button>
                                                <div className="dropdown-menu" aria-labelledby="btnGroupDrop1">
                                                    <a className="dropdown-item" onClick={self.resetPermissions.bind(self, row)}>Reset settings</a>
                                                    <a className={row.Deleteable ? 'dropdown-item' : 'hidden'} onClick={self.deleteNamespace.bind(self, row)}>Delete</a>
                                                </div>
                                            </div>
                                        );
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

export default onClickOutside(K8sNamespace);

