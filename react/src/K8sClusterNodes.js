import React, { Component } from 'react';
import $ from 'jquery';

class K8sClusterNodes extends Component {
    constructor(props) {
        super(props);

        this.state = {
            globalMessage: '',
            searchValue: '',
            nodes: [],
        };

        setInterval(() => {
            this.refresh()
        }, 10000);
    }

    loadNodes() {
        $.get({
            url: '/api/cluster/nodes'
        }).done((data) => {
            this.setState({
                nodes: data
            });
        });
    }

    componentDidMount() {
        this.loadNodes();
    }

    refresh() {
        this.loadNodes();
    }

    handleSearchChange(event) {
        this.setState({
            searchValue: event.target.value
        });
    }

    getNodes() {
        let ret = [];
        if (this.state.searchValue !== "") {
            let term =this.state.searchValue;
            ret = this.state.nodes.filter((row) => {
                if (row.Name.includes(term)) {
                    return true;
                }

                return false;
            });
        } else {
            ret = this.state.nodes;
        }

        ret = ret.sort(function(a,b) {
            return a.Name >= b.Name;
        });

        return ret;
    }

    render() {
        let nodes = this.getNodes();
        if (nodes) {
            return (
                <div>
                    <div className="container-toolbar-main">
                        <div className={this.state.globalMessage === '' ? 'alert alert-success invisible' : 'alert alert-success'}>{this.state.globalMessage}</div>
                        <input type="text" className="form-control search-input" placeholder="Search" value={this.state.searchValue} onChange={this.handleSearchChange.bind(this)} />
                        <div class="clearfix"></div>
                    </div>
                    <table className="table table-hover table-sm">
                        <thead>
                        <tr>
                            <th>Name</th>
                            <th>System</th>
                            <th>Version</th>
                            <th>Created</th>
                            <th>Status</th>
                        </tr>
                        </thead>
                        <tbody>
                        {nodes.map((row) =>
                            <tr key={row.Name} className={row.Role === 'master' ? 'table-warning' : null}>
                                <td>
                                    <span
                                        className={row.Role === 'master' ? 'badge badge-danger' : 'badge badge-primary'}>{row.Role}</span> {row.Name}<br/>
                                    <span className="badge badge-info">{row.SpecArch}</span>
                                    <span className="badge badge-info">{row.SpecOS}</span>
                                    <span className="badge badge-secondary">Region {row.SpecRegion}</span>
                                    <span className="badge badge-secondary">Zone {row.SpecZone}</span>
                                </td>
                                <td>
                                    <small>
                                        {row.SpecInstance}<br/>
                                        CPU: {row.SpecMachineCPU}<br/>
                                        MEM: {row.SpecMachineMemory}<br/>
                                    </small>
                                </td>
                                <td>{row.Version}</td>
                                <td><div title={row.Created}>{row.CreatedAgo}</div></td>
                                <td>
                                    <span
                                        className={row.Status === 'Ready' ? 'badge badge-success' : 'badge badge-warning'}>{row.Status}</span>
                                </td>
                            </tr>
                        )}
                        </tbody>
                    </table>
                </div>);
        }
    }
}

export default K8sClusterNodes;

