import { Component } from 'react';
import $ from 'jquery';

class BaseComponent extends Component {

    setInputFocus() {
        setTimeout( () => {
            $(":input:text:visible:enabled").first().focus();
        }, 500);
    }

    handleXhr(jqxhr) {
        jqxhr.fail((jqxhr) => {
            if (jqxhr.status === 401) {
                this.setState({
                    globalError: "Login expired, please reload",
                    isStartup: false
                });
            } else if (jqxhr.responseJSON && jqxhr.responseJSON.Message) {
                this.setState({
                    globalError: jqxhr.responseJSON.Message,
                    isStartup: false
                });
            }
        });
    }

}
export default BaseComponent;
