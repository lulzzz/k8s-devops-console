import React, { Component } from 'react';
import $ from 'jquery';

class K8sNamespaceModalDelete extends Component {
    constructor(props) {
        super(props);

        this.state = {
            message: "",
            buttonState: "",
            buttonText: "Delete namespace",
            confirmNamespace: ""
        };
    }

    deleteNamespace() {
        if (!this.props.namespace) {
            return
        }

        let oldButtonText = this.state.buttonText;
        this.setState({
            buttonState: "disabled",
            buttonText: "Deleting...",
            message: ""
        });

        $.ajax({
            type: 'DELETE',
            url: "/api/namespace/" + encodeURI(this.props.namespace.Name)
        }).done(() => {
            this.setState({
                confirmNamespace: ""
            });

            if (this.props.callback) {
                this.props.callback(this.props.namespace.Name)
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

    componentWillReceiveProps(nextProps) {
        if (!this.props.namespace || this.props.namespace.Name !== nextProps.namespace.Name) {
            this.setState({
                confirmNamespace: ""
            });
        }
    }

    handleConfirmNamespace(event) {
        this.setState({
            confirmNamespace: event.target.value
        });
    }

    renderButtonState() {
        if (this.state.buttonState !== "") {
            return this.state.buttonState;
        }

        if (this.state.confirmNamespace !== this.props.namespace.Name) {
            return "disabled";
        }
    }

    render() {
        return (
            <div>
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
                                <div className="row">
                                    <div className="col">
                                        <div className={this.state.message === '' ? null : 'alert alert-danger'}>{this.state.message}</div>
                                        Do you really want to delete namespace <strong className="k8s-namespace">{this.props.namespace.Name}</strong>?
                                    </div>
                                </div>
                                <div className="row">
                                    <div className="col">
                                        <input type="text" id="inputNsDeleteConfirm" className="form-control" placeholder="Enter namespace for confirmation" required value={this.state.confirmNamespace} onChange={this.handleConfirmNamespace.bind(this)} />
                                    </div>
                                </div>
                            </div>
                            <div className="modal-footer">
                                <button type="button" className="btn btn-primary bnt-k8s-namespace-cancel" data-dismiss="modal">Cancel</button>
                                <button type="button" className="btn btn-secondary bnt-k8s-namespace-delete" disabled={this.renderButtonState()} onClick={this.deleteNamespace.bind(this)}>{this.state.buttonText}</button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        );
    }
}

export default K8sNamespaceModalDelete;

