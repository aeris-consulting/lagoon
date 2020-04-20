import {actions, getters, mutations} from '../../../src/store/datasource.module';
import {testAction} from './store.test.helper';
import {SET_DATASOURCES} from '../../../src/store/mutations.type'
import {datasourceResponse} from '../data/api-data'
import {DatasourcesService} from '../../../src/services/api.service'

jest.mock('../../../src/services/api.service');

describe('mutations', () => {
    it('UNSELECT_NODE', () => {
        // given
        const state = {
            selectedNodes: [
                {fullPath: 'path01'},
                {fullPath: 'path02'}
            ]
        }
        const nodeToUnSelect = {fullPath: 'path01'}

        // when
        mutations.UNSELECT_NODE(state, nodeToUnSelect)

        // then
        expect(state.selectedNodes.length).toEqual(1)
        expect(state.selectedNodes[0].fullPath).toEqual('path02')
    })
})

describe('actions', () => {
    it('FETCH_DATASOURCE', (done) => {
        const state = {
            datasources: []
        }
        DatasourcesService.getDatasources.mockResolvedValue(datasourceResponse);

        testAction(actions.FETCH_DATASOURCE, null, state, [
            {type: SET_DATASOURCES, payload: datasourceResponse.datasources}
        ], done)
    })
})

describe('getters', () => {
    it('getSelected', () => {
        // given
        const state = {
            datasources: datasourceResponse.datasources,
            selectedDatasourceId: datasourceResponse.datasources[0].id
        }

        // when
        const selectedDatasource = getters.getSelected(state)

        // then
        expect(selectedDatasource).toEqual(datasourceResponse.datasources[0])
    })
})
  