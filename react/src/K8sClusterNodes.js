import React from 'react';
import $ from 'jquery';

import BaseComponent from './BaseComponent';
import Spinner from './Spinner';

class K8sClusterNodes extends BaseComponent {
    constructor(props) {
        super(props);

        this.state = {
            isStartup: true,
            globalMessage: '',
            globalError: '',
            searchValue: '',
            nodes: [],
        };

        setInterval(() => {
            this.refresh()
        }, 10000);
    }

    loadNodes() {
        let jqxhr = $.get({
            url: '/api/cluster/nodes'
        }).done((jqxhr) => {
            this.setState({
                nodes: jqxhr,
                globalError: '',
                isStartup: false
            });
        });

        this.handleXhr(jqxhr);
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
            let term = this.state.searchValue.replace(/[.?*+^$[\]\\(){}|-]/g, "\\$&");
            let re = new RegExp(term, "i");

            ret = this.state.nodes.filter((row) => {
                if (row.Name.search(re) !== -1) return true;
                if (row.SpecArch.search(re) !== -1) return true;
                if (row.SpecRegion.search(re) !== -1) return true;
                if (row.SpecOS.search(re) !== -1) return true;
                if (row.SpecZone.search(re) !== -1) return true;
                if (row.SpecInstance.search(re) !== -1) return true;
                if (row.SpecMachineCPU.search(re) !== -1) return true;
                if (row.SpecMachineMemory.search(re) !== -1) return true;
                if (row.Version.search(re) !== -1) return true;

                return false;
            });
        } else {
            ret = this.state.nodes;
        }

        ret = ret.sort(function(a,b) {
            if(a.Name < b.Name) return -1;
            if(a.Name > b.Name) return 1;
            return 0;
        });

        return ret;
    }

    render() {
        if (this.state.globalError) {
            return (
                <div className="alert alert-danger">{this.state.globalError}</div>
            )
        }

        let nodes = this.getNodes();
        if (nodes) {
            return (
                <div>
                    <Spinner active={this.state.isStartup}/>
                    <div className="container-toolbar-main">
                        <div className="floating-message">
                            <div className={this.state.globalMessage === '' ? 'alert alert-success invisible' : 'alert alert-success'}>{this.state.globalMessage}</div>
                        </div>
                        <input type="text" className="form-control search-input" placeholder="Search" value={this.state.searchValue} onChange={this.handleSearchChange.bind(this)} />
                        <div className="clearfix"></div>
                    </div>
                    <table className="table table-hover table-sm">
                        <thead>
                        <tr>
                            <th>Name</th>
                            <th>Network</th>
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
                                    <span className={row.Role === 'master' ? 'badge badge-danger' : 'badge badge-primary'}>{row.Role}</span> {row.Name}<br/>
                                    <span className="badge badge-info">{row.SpecArch}</span>
                                    <span className="badge badge-info">{row.SpecOS}</span>
                                    <span className="badge badge-secondary">Region {row.SpecRegion}</span>
                                    <span className="badge badge-secondary">Zone {row.SpecZone}</span>
                                </td>
                                <td>{row.InternalIp}</td>
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
                                        className={row.Status === 'Ready' ? 'badge badge-success' : 'badge badge-warning'}>{row.Status !== '' ? row.Status  : "unknown"}</span>
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

