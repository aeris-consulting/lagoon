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

        if (process !== null && process.env !== null && process.env.VUE_APP_API_SCHEME && process.env.VUE_APP_API_URL) {
            this.apiRoot = process.env.VUE_APP_API_SCHEME + '://' + process.env.VUE_APP_API_URL;
            if (process.env.VUE_APP_API_SCHEME == 'https:') {
                this.wsRoot = 'wss://' + process.env.VUE_APP_API_URL;
            } else {
                this.wsRoot = 'ws://' + process.env.VUE_APP_API_URL;
            }
        } else {
            this.apiRoot = '..';
            if (location.protocol == 'https:') {
                this.wsRoot = 'wss://' + location.hostname + ':' + location.port;
            } else {
                this.wsRoot = 'ws://' + location.hostname + ':' + location.port;
            }
        }
        // eslint-disable-next-line
        console.log(this);
    }

    listEntrypoints(entrypointPrefix, minLevel, maxLevel, completeAction, onError) {
        let receivedValues = [];
        let actualFilter;
        let overallFilter = ('*' + this.filter + '*').replace(/[*]+/g, '*');

        if (!entrypointPrefix) {
            actualFilter = ('*' + this.filter + '*').replace(/[*]+/g, '*');
        } else {
            let entrypointRegex = new RegExp(entrypointPrefix);
            entrypointPrefix = entrypointPrefix + ':*';
            let overallRegex = new RegExp(overallFilter
                .replace(/^[*]+/g, '')
                .replace(/[*]+$/g, '')
                .replace(/[*]+/g, '.*')
            );

            if (overallRegex.test(entrypointPrefix)) {
                actualFilter = entrypointPrefix;
            } else if (entrypointRegex.test(overallFilter
                .replace(/^[*]+/g, '')
                .replace(/[*]+$/g, '')
            )) {
                actualFilter = overallFilter;
            } else {
                // TODO Complex case.
                actualFilter = entrypointPrefix + ','
                    + overallFilter
                        .replace(/[*]+/g, '.*')
                        .replace(/[*]+/g, '*');
            }
        }

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
                this.addError(e);
                if (onError) {
                    onError(e);
                }
            });
    }

    addError(e) {
        if (e.response && e.response.data && e.response.data.error) {
            this.errors = [e.response.data.error];
        } else {
            this.errors = [e];
        }
    }

    refreshNodeDetails(node, callback) {
        let fullName = node.getFullName();
        let self = this;
        axios.get(this.apiRoot + '/data/' + this.id + '/entrypoint/' + fullName + '/info', {format: 'json'})
            .then(response => {
                if (response.status === 200) {
                    node.info = response.data;
                    axios.get(self.apiRoot + '/data/' + self.id + '/entrypoint/' + fullName + '/content', {format: 'json'})
                        .then(response => {
                            if (response.status === 200) {
                                node.content = response.data;
                                if (callback) {
                                    callback();
                                }
                            } else if (response.status === 202) {
                                let receivedValues = [];
                                let socket = new WebSocket(this.wsRoot + response.data.link);
                                socket.onopen = () => {
                                    socket.onmessage = ({data}) => {
                                        let jsonData = JSON.parse(data);
                                        if (jsonData.size) {
                                            receivedValues = receivedValues.concat(jsonData.data);
                                        } else {
                                            setTimeout(() => {
                                                node.content = {
                                                    length: receivedValues.length,
                                                    data: receivedValues
                                                };
                                                if (callback) {
                                                    callback();
                                                }
                                            }, 0);
                                        }
                                    };
                                };
                            }
                        })
                        .catch(e => {
                            self.addError(e);
                        });
                }
            })
            .catch(e => {
                this.addError(e);
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
                this.addError(e);
            });
    }

    deleteEntrypointChildren(node) {
        let fullName = node.getFullName();

        axios.delete(this.apiRoot + '/data/' + this.id + '/entrypoint/' + fullName + '/children', {format: 'json'})
            .then(response => {
                if (response.status === 202) {
                    let socket = new WebSocket(this.wsRoot + response.data.link);
                    socket.onopen = () => {
                        socket.onmessage = () => {
                            setTimeout(() => {
                                node.parent.component.refresh();
                                this.unselectNode(node);
                            });
                        };
                    };
                }
            })
            .catch(e => {
                this.addError(e);
            });
    }

    selectNode(node) {
        this.selectedNodes = [];
        setTimeout(() => {
            this.selectedNodes.push(node);
        }, 0);
    }

    // eslint-disable-next-line
    unselectNode(node) {
        this.selectedNodes = [];
    }
}