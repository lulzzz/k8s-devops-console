import React, { Component } from 'react';
import { BrowserRouter as Router, Route } from 'react-router-dom';
import K8sClusterNodes from './K8sClusterNodes';
import K8sNamespace from './K8sNamespace';
import $ from "jquery";

class App extends Component {
    constructor(props) {
        super(props);

        this.state = {
            loggedIn: false,
            username: "",
            password: "",
            buttonState: "disabled"
        };
    }

    handleChangeUsername(event) {
        if (event.target.value !== "") {
            this.setState({buttonState: ""});
        } else {
            this.setState({buttonState: "disabled"});
        }
        this.setState({username: event.target.value});
    }

    handleChangePassword(event) {
        if (event.target.value !== "") {
            this.setState({buttonState: ""});
        } else {
            this.setState({buttonState: "disabled"});
        }
        this.setState({password: event.target.value});
    }

    handleLogin() {
        $.ajax({
            type: 'POST',
            url: "/api/_login",
            data: {
                username: this.state.username,
                password: this.state.password
            }
        }).done((data) => {
            this.setState({loggedIn: true});
        });
    }

    render() {
        return (
            <Router>
                <div>
                    <Route path="/cluster" component={K8sClusterNodes} />
                    <Route path="/namespace" component={K8sNamespace} />
                </div>
            </Router>
        )
    }
}

export default App;
