import {
    FETCH_DATASOURCE,
    FETCH_ENTRY_POINTS,
    SELECT_DATASOURCE,
    SELECT_NODE,
    FETCH_NODE_DETAILS,
    DELETE_NODE
} from './actions.type';
import {
    SET_DATASOURCE,
    SET_SELECTED_DATASOURCE,
    SET_ENTRY_POINTS,
    SET_SELECTED_NODE,
    UNSELECT_NODE,
    ADD_ERROR
} from './mutations.type';

import { DatasourcesService } from '../services/api.service'

const initialState = {
    selectedDatasourceId: null,
    selectedNodes: [],
    datasources: [],
    entryPoints: []
}

const state = { ...initialState }

const getters = {
    getSelected: (state) => () => {
        return state.datasources.find(datasource => datasource.id === state.selectedDatasourceId)
    }
}

export const actions = {
    [FETCH_DATASOURCE](context) {
        return DatasourcesService.getDatasources().then((data) => {
            context.commit(SET_DATASOURCE, data.datasources);
            return data.datasources;
        }).catch(e => {
            context.commit(ADD_ERROR, e);
        })
    },
    [FETCH_NODE_DETAILS](context, node) {
        return DatasourcesService.getNodeDetails(context.getters.getSelected(), node)
            .then((details) => {
                return details
            });
    },
    [SELECT_DATASOURCE](context, datasourceId) {
        context.commit(SET_SELECTED_DATASOURCE, datasourceId);
        return datasourceId;
    },
    [DELETE_NODE](context, node) {
        return DatasourcesService.deleteEntrypoint(context.state.selectedDatasourceId, node.fullPath)
            .then(() => {
                context.commit(UNSELECT_NODE, node);
            })
            .catch(() => {

            });
    },
    [SELECT_NODE](context, node) {
        context.commit(SET_SELECTED_NODE, node);
        return node;
    },
    [FETCH_ENTRY_POINTS](context, request) {
        const {filter, minLevel, maxLevel} = request;
        let { entrypointPrefix } = request;
        let actualFilter;
        let overallFilter = ('*' + filter + '*').replace(/[*]+/g, '*');

        if (!entrypointPrefix) {
            actualFilter = ('*' + filter + '*').replace(/[*]+/g, '*');
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
            })
        })
        
    }
}

export const mutations = {
    [SET_DATASOURCE](state, datasources) {
        state.datasources = datasources;
    },
    [SET_ENTRY_POINTS](state, entryPoints) {
        state.entryPoints = entryPoints;
    },
    [SET_SELECTED_DATASOURCE](state, selectedDatasourceId) {
        state.selectedDatasourceId = selectedDatasourceId;
    },
    [SET_SELECTED_NODE](state, node) {
        state.selectedNodes = [node];
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