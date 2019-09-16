import { shallowMount } from '@vue/test-utils'
import axios from 'axios'
import EntrypointContent from '../../../src/components/EntrypointContent.vue'
import DataSource from '../../../src/models/DataSource.js'
import Node from '../../../src/models/Node.js'

jest.mock('axios');

beforeEach(() => {
  axios.get.mockImplementation((url, ...otherParams) => {
    if (url.endsWith('info')) {
      return Promise.resolve({
        status: 200,
        data: {
          length: 3,
          type: 'VALUE'
        }
      })
    } else if (url.endsWith('content')) {
      return Promise.resolve({
        status: 200,
        data: {
          data: ['value1', 'value2'],
          size: 1
        }
      })
    }
  });
});

test('EntrypointContent', (done) => {
  const nodeName = 'testNodeName'
  const wrapper = shallowMount(EntrypointContent, {
    propsData: {
      dataSource: new DataSource('dataSourceId', ''),
      node: new Node(nodeName, 1, true)
    }
  })

  wrapper.vm.$nextTick(() => {
    expect(wrapper.vm.node.getFullName()).toEqual(nodeName)
    expect(wrapper.vm.node.info.type).toEqual('VALUE')
    done();
  });
})