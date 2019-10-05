import axios from 'axios';

var _ = require('lodash');

export default class DataSource {

    constructor(id, filter) {
        this.id = id;
        this.level = 0;
        this.filter = filter;
        this.status = null;
        this.errors = [];
        this.selectedNodes = [];
        this.readonly = false;

        if (!_.isNil(process) && !_.isNil(process.env) && !_.isNil(process.env.VUE_APP_API_BASE_URL) && !_.isNil(process.env.VUE_APP_WS_BASE_URL)) {
            this.apiRoot = process.env.VUE_APP_API_BASE_URL;
            this.wsRoot = process.env.VUE_APP_WS_BASE_URL;
        } else {
            this.apiRoot = location.pathname + '..';
            if (location.protocol == 'https:') {
                this.wsRoot = 'wss://';
            } else {
                this.wsRoot = 'ws://';
            }
            this.wsRoot += location.hostname + ':' + location.port + this.apiRoot
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
                actualFilter = overallFilter + ','
                    + entrypointPrefix
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

    refreshNodeDetails(node) {
        return new Promise((resolve, reject) => {
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
                                    resolve();
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
                                                    resolve();
                                                }, 0);
                                            }
                                        };
                                    };
                                }
                            })
                            .catch(e => {
                                self.addError(e);
                                reject();
                            });
                    }
                })
                .catch(e => {
                    this.addError(e);
                    reject();
                });
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