import {
    DELETE_CHILDREN_NODE,
    DELETE_NODE,
    FETCH_DATASOURCE,
    FETCH_ENTRY_POINTS,
    FETCH_NODE_DETAILS,
    SELECT_DATASOURCE,
    SELECT_NODE
} from './actions.type';
import {
    ADD_ERROR,
    SET_DATASOURCES,
    SET_ENTRY_POINTS,
    SET_SELECTED_DATASOURCE,
    ADD_SELECTED_NODE,
    UNSELECT_NODE,
    SET_SELECTED_NODES,
    NODE_DELETED
} from './mutations.type';

import {DatasourcesService} from '../services/api.service'
import FilterHelper from '../helpers/filterHelper'

const initialState = {
    selectedDatasourceId: null,
    selectedNodes: [],
    datasources: [],
    entryPoints: []
}

const state = {...initialState}

export const getters = {
    getSelected(state) {
        return state.datasources.find(datasource => datasource.id === state.selectedDatasourceId)
    },
}

export const actions = {
    [FETCH_DATASOURCE](context) {
        return DatasourcesService.getDatasources().then((data) => {
            context.commit(SET_DATASOURCES, data.datasources);
            return data.datasources;
        }).catch(e => {
            context.commit(ADD_ERROR, e);
        })
    },
    [FETCH_NODE_DETAILS](context, node) {
        return DatasourcesService.getNodeDetails(context.getters.getSelected, node)
            .then((details) => {
                return details
            }).catch(e => {
                context.commit(ADD_ERROR, e);
            });
    },
    [SELECT_DATASOURCE](context, datasourceId) {
        context.commit(SET_SELECTED_DATASOURCE, datasourceId);
        return datasourceId;
    },
    [DELETE_NODE](context, node) {
        return DatasourcesService.deleteEntrypoint(context.state.selectedDatasourceId, node.fullPath)
            .then(() => {
                context.commit(NODE_DELETED, node);
                context.commit(UNSELECT_NODE, node);
            })
            .catch((e) => {
                context.commit(ADD_ERROR, e);
            });
    },
    [DELETE_CHILDREN_NODE](context, node) {
        return DatasourcesService.deleteEntrypointChildren(context.state.selectedDatasourceId, node.fullPath)
            .catch((e) => {
                context.commit(ADD_ERROR, e);
            });
    },
    [SELECT_NODE](context, node) {
        context.commit(ADD_SELECTED_NODE, node);
        return node;
    },
    [FETCH_ENTRY_POINTS](context, request) {
        const {filter, minLevel, maxLevel} = request;
        const {entrypointPrefix} = request;
        const actualFilter = FilterHelper.transformFilter(entrypointPrefix, filter)
        return new Promise((resolve, reject) => {
            DatasourcesService.listEntryPoints({
                id: context.state.selectedDatasourceId,
                filter: actualFilter,
                minLevel,
                maxLevel
            }).then(response => {
                if (response.status === 202) {
                    DatasourcesService.getEntryPointsFromWebsocket(response.data.link).then(data => {
                        resolve(data)
                    }).catch(e => {
                        context.commit(ADD_ERROR, e);
                    })
                } else {
                    reject()
                }
            }).catch(e => {
                context.commit(ADD_ERROR, e);
                reject()
            })
        })

    }
}

export const mutations = {
    [SET_DATASOURCES](state, datasources) {
        state.datasources = datasources;
    },
    [SET_ENTRY_POINTS](state, entryPoints) {
        state.entryPoints = entryPoints;
    },
    [SET_SELECTED_DATASOURCE](state, selectedDatasourceId) {
        state.selectedDatasourceId = selectedDatasourceId;
    },
    [ADD_SELECTED_NODE](state, node) {
        if (!state.selectedNodes.some(n => n.fullPath === node.fullPath)) {
            state.selectedNodes.push(node)
        }
    },
    [SET_SELECTED_NODES](state, nodes) {
        state.selectedNodes = [...nodes]
    },
    [UNSELECT_NODE](state, node) {
        state.selectedNodes = state.selectedNodes.filter(n => n.fullPath !== node.fullPath);
    }
}

export default {
    state,
    actions,
    mutations,
    getters
}