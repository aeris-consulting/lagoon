import Vuex from "vuex"
import { shallowMount, createLocalVue } from "@vue/test-utils"
import DataSourceComponent from "../../../src/components/DataSource"



const localVue = createLocalVue()
localVue.use(Vuex)

function createNode(fullPath) {
  return {
    fullPath
  }
}

const mutations = {
  SET_SELECTED_NODES: jest.fn()
}

const store = new Vuex.Store({
  modules: {
    datasource: {
      state: {
        selectedDatasourceId: "single",
        selectedNodes: [
          createNode('a:001'),
          createNode('a:002'),
          createNode('a:003')
        ],
        datasources: [{
          "id": "single",
          "vendor": "redis",
          "name": "Single",
          "description": "",
          "readonly": false
        }],
        entryPoints: []
      },
      actions: {
        'FETCH_DATASOURCE': jest.fn()
      },
      mutations
    }
  }
})

describe("DataSourceComponent", () => {
  let wrapper
  beforeEach(() => {
    wrapper = shallowMount(DataSourceComponent, {
      store, 
      localVue
    })
  })

  it("Should get value from state", () => {
    expect(wrapper.vm.datasources[0].id).toBe('single')
  })

  it("close all but unpinned", () => {
    // given
    wrapper.vm.togglePin(createNode('a:001'))
    wrapper.vm.togglePin(createNode('a:002'))

    // when
    wrapper.vm.closeAllButPinned()

    // then
    expect(mutations.SET_SELECTED_NODES)
    .toHaveBeenCalledWith(expect.anything(), [
      createNode('a:001'),
      createNode('a:002'),
    ])
  })

  it("close others nodes", () => {
    // given
    const nodeToKeep = createNode('a:001')

    // when
    wrapper.vm.closeOthers(nodeToKeep)

    // then
    expect(mutations.SET_SELECTED_NODES)
    .toHaveBeenCalledWith(expect.anything(), [
      createNode('a:001')
    ])
  })

  it("close others nodes should keep pinned node", () => {
    // given
    const nodeToKeep = createNode('a:001')
    wrapper.vm.togglePin(createNode('a:002'))

    // when
    wrapper.vm.closeOthers(nodeToKeep)

    // then
    expect(mutations.SET_SELECTED_NODES)
    .toHaveBeenCalledWith(expect.anything(), [
      createNode('a:001'),
      createNode('a:002'),
    ])
  })
})
