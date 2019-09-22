import axios from 'axios'
import { Server } from 'mock-socket';
import DataSource from '../../../src/models/DataSource.js'
import Node from '../../../src/models/Node.js'

jest.mock('axios')
let dataSource = null
let mockWebSocketServer = null
const dataToSend = ['value1', 'value2', 'value3', 'value4']

beforeEach(() => {
    const webSocketLink = 'ws.link'
    dataSource = new DataSource("source-id", "filter")
    mockWebSocketServer = new Server(dataSource.wsRoot + webSocketLink)
  
    mockWebSocketServer.on('connection', socket => {

        // send first part
        socket.send(
            JSON.stringify(
                {
                    size: 2,
                    data: dataToSend.slice(0, 2)
                }
            )
        );

        // send second part
        socket.send( 
            JSON.stringify(
                {
                    size: 2,
                    data: dataToSend.slice(2)
                }
            )
        );

        socket.send(
            JSON.stringify({})
        );
    });

    axios.get.mockImplementation((url) => {
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
            status: 202,
            data: {
                link: webSocketLink
            }
        })
    }
    });
});

test('refreshNodeDetails get websocket link', (done) => {
    const node = new Node('name', 1, true)
    dataSource.refreshNodeDetails(node).then(() => {
        expect(node.content).toEqual({
            length: 4,
            data: dataToSend
        })
        done()
    })
});
