import axios from 'axios';

var _ = require('lodash');

export default class DataSource {

    constructor(id, readonly, filter) {
        this.id = id;
        this.readonly = readonly;
        this.level = 0;
        this.filter = filter;
        this.status = null;
        this.errors = [];
        this.selectedNodes = [];

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

    listEntrypoints(entrypointPrefix, minLevel, maxLevel, completeAction, onError, onClose) {
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
                            if (jsonData.size > 0) {
                                receivedValues = receivedValues.concat(jsonData.data);
                            } else {
                                setTimeout(() => {
                                    // eslint-disable-next-line
                                    console.log("Closing the websocket");
                                    socket.close(1000, "End of data");
                                    // eslint-disable-next-line
                                    console.log(receivedValues)
                                    console.log("Count of received values: %d", receivedValues.length);
                                    completeAction(receivedValues);
                                }, 0);
                            }
                        };
                        socket.onerror = (e) => {
                            setTimeout(() => {
                                // eslint-disable-next-line
                                console.log("The websocket got an error: ", e);
                                onError(e);
                            }, 0);
                        };
                        socket.onclose = () => {
                            setTimeout(() => {
                                // eslint-disable-next-line
                                console.log("The websocket is closed");
                                onClose();
                            }, 0);
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
        if (this.readonly) {
            this.addError('This data source can only be read');
            return;
        }

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

    deleteEntrypointChildren(node, component) {
        if (this.readonly) {
            this.addError('This data source can only be read');
            return;
        }

        let fullName = node.getFullName();
        let self = this;

        axios.delete(this.apiRoot + '/data/' + this.id + '/entrypoint/' + fullName + '/children', {format: 'json'})
            .then(response => {
                if (response.status === 202) {
                    let socket = new WebSocket(this.wsRoot + response.data.link);
                    socket.onopen = () => {
                        socket.onmessage = ({data}) => {
                            self.addError(JSON.parse(data));
                        };
                    };
                    socket.onclose = () => {
                        setTimeout(() => {
                            this.unselectNode(node);
                            component.loading = false;

                            let parent = node.parent;
                            if (parent.component == null) {
                                // A root was deleted.
                                component.$destroy();
                                component.$el.parentNode.removeChild(component.$el);
                            } else {
                                while (parent != null && parent.component != null) {
                                    parent.component.refresh();
                                    parent = parent.parent;
                                }
                            }
                        });
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

    async getClusterNodes() {
        return axios.post(this.apiRoot + '/data/' + this.id + '/command', {args: ['cluster', 'nodes']})
            .then(response => {
                const clusterNodes = response.data.data.split(/\n/)
                    .map(nodeInfoString => nodeInfoString.split(' '))
                    .filter(infos => infos.length >= 3)
                    .map(infos => {
                        return {
                            id: infos[0],
                            ip: infos[1].split('@')[0],
                            role: infos[2].replace('myself,', '')
                        }
                    });
                return clusterNodes;
            }).catch(() => {
                // not a cluster, do nothing
            });
    }

    async executeCommand(commands, nodeId) {
        return axios.post(this.apiRoot + '/data/' + this.id + '/command', {args: commands, nodeId: nodeId})
            .then(response => {
                return response.data;
            }).catch(e => {
                return Promise.reject(e.response.data.error)
            });
    }
}