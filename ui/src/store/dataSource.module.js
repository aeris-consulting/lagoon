import {
    SET_DATASOURCE,
    FETCH_DATASOURCE,
    FETCH_ENTRY_POINTS,
    SET_ENTRY_POINTS,
    SELECT_DATASOURCE,
    SELECT_NODE,
    SET_SELECTED_NODE,
    SET_SELECTED_DATASOURCE
} from './actions.type';

import { DatasourcesService } from '../services/api.service'

const initialState = {
    selectedDatasourceId: null,
    selectedNode: null,
    datasources: [],
    entryPoints: []
}

const state = { ...initialState }

const getters = {
    getSelectedDatasource: (state) => () => {
        return state.datasources.find(datasource => datasource.id === state.selectedDatasourceId)
    }
}

export const actions = {
    async [FETCH_DATASOURCE](context) {
        const { data } = await DatasourcesService.getDatasources();
        context.commit(SET_DATASOURCE, data.datasources);
        return data.datasources;
    },
    [SELECT_DATASOURCE](context, datasourceId) {
        context.commit(SET_SELECTED_DATASOURCE, datasourceId);
        return datasourceId;
    },
    [SELECT_NODE](context, node) {
        context.commit(SET_SELECTED_NODE, node);
        return datasourceId;
    },
    async [FETCH_ENTRY_POINTS](context, request) {
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

        const response = await DatasourcesService.listEntryPoints({
            id: context.state.selectedDatasourceId,
            filter: actualFilter,
            minLevel,
            maxLevel
        });
        
        if (response.status === 202) {
            const data = await DatasourcesService.getEntryPointsFromWebsocket(response.data.link)
            console.log(data)
            return data;
        }
        return Promise.reject()
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
        state.selectedDatasourceId = selectedDatasourceId
    },
    [SET_SELECTED_NODE](state, node) {
        state.selectedNode = node
    }
}

export default {
    state,
    actions,
    mutations,
    getters
}