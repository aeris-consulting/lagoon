import axios from 'axios';

export default class DataSource {

    constructor(id, filter) {
        this.id = id;
        this.level = 0;
        this.filter = filter;
        this.status = null;
        this.errors = [];
        this.selectedNodes = [];
        this.readonly = false;

        if (process.env.VUE_APP_API_SCHEME && process.env.VUE_APP_API_URL) {
            this.apiRoot = process.env.VUE_APP_API_SCHEME + '://' + process.env.VUE_APP_API_URL;
            if (process.env.VUE_APP_API_SCHEME == 'https') {
                this.wsRoot = 'wss://' + process.env.VUE_APP_API_URL;
            } else {
                this.wsRoot = 'ws://' + process.env.VUE_APP_API_URL;
            }
        } else {
            this.apiRoot = '..';
            if (location.protocol == 'https') {
                this.wsRoot = 'wss://' + location.hostname + ':' + location.port;
            } else {
                this.wsRoot = 'ws://' + location.hostname + ':' + location.port;
            }
        }
        // eslint-disable-next-line
        console.log(this);
    }

    listEntrypoints(filter, minLevel, maxLevel, completeAction, onError) {
        let receivedValues = [];
        let actualFilter = filter ? filter : this.filter ? '*' + this.filter + '*' : '*';

        axios.get(this.apiRoot + '/data/' + this.id + '/entrypoint?min='
            + minLevel + '&max=' + maxLevel + '&filter=' + actualFilter, {format: 'json'})
            .then(response => {
                if (response.status === 202) {
                    let socket = new WebSocket(this.wsRoot + response.data.link);
                    socket.onopen = () => {
                        socket.onmessage = ({data}) => {
                            let jsonData = JSON.parse(data);
                            if (jsonData.size) {
                                receivedValues = receivedValues.concat(jsonData.data);
                            } else {
                                setTimeout(() => {
                                    socket.close();
                                    socket = null;
                                    // eslint-disable-next-line
                                    console.log("Count of received values: %d", receivedValues.length);
                                    completeAction(receivedValues);
                                }, 0);
                            }
                        };
                    };
                }
            })
            .catch(e => {
                this.errors.push(e);
                if (onError) {
                    onError(e);
                }
            });
    }

    refreshNodeDetails(node) {
        let fullName = node.getFullName();

        axios.get(this.apiRoot + '/data/' + this.id + '/entrypoint/' + fullName + '/info', {format: 'json'})
            .then(response => {
                if (response.status === 200) {
                    node.info = response.data;
                    node.contentComponent.lastRefresh = new Date();
                }
            })
            .catch(e => {
                this.errors.push(e)
            });

        axios.get(this.apiRoot + '/data/' + this.id + '/entrypoint/' + fullName + '/content', {format: 'json'})
            .then(response => {
                if (response.status === 200) {
                    node.content = response.data;
                    node.contentComponent.lastRefresh = new Date();
                }
            })
            .catch(e => {
                this.errors.push(e)
            });
    }

    deleteEntrypoint(node) {
        let fullName = node.getFullName();

        axios.delete(this.apiRoot + '/data/' + this.id + '/entrypoint/' + fullName, {format: 'json'})
            .then(response => {
                if (response.status === 200) {
                    node.parent.component.refresh();
                    this.unselectNode(node);
                }
            })
            .catch(e => {
                this.errors.push(e)
            });
    }

    deleteEntrypointChildren(node) {
        let fullName = node.getFullName();

        axios.delete(this.apiRoot + '/data/' + this.id + '/entrypoint/' + fullName + '/children', {format: 'json'})
            .then(response => {
                if (response.status === 202) {
                    let socket = new WebSocket(this.wsRoot + response.data.link);
                    socket.onopen = () => {
                        socket.onmessage = ({data}) => {
                            setTimeout(() => {
                                node.parent.component.refresh();
                                this.unselectNode(node);
                            });
                        };
                    };
                }
            })
            .catch(e => {
                this.errors.push(e)
            });
    }

    selectNode(node) {
        this.selectedNodes = [];
        setTimeout(() => {
            this.selectedNodes.push(node);
        }, 0);

    }

    unselectNode(node) {
        this.selectedNodes = [];
    }
}