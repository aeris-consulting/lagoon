import axios from 'axios'
import {Server} from 'mock-socket';
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
        });
        done()
    })
});

test.only('fetching cluster nodes information', (done) => {
    axios.get.mockImplementation(() => {
        return Promise.resolve({
            status: 200,
            data: {
                infos: {
                    nodes: [
                        {
                            id: "my-id-1",
                            server: "my-server-1",
                            name: "my-name-1",
                            role: "my-role-1",
                            masters: ["my-master-1", "my-master-2"]
                        },
                        {
                            id: "my-id-2",
                            server: "my-server-2",
                            name: "my-name-2",
                            role: "my-role-2",
                            masters: ["my-master-2", "my-master-1"]
                        }
                    ]
                }
            }
        })
    });
    dataSource.getClusterNodes().then((data) => {
        expect(data[0].id).toEqual('my-id-1');
        expect(data[0].server).toEqual('my-server-1');
        expect(data[0].name).toEqual('my-name-1');
        expect(data[0].role).toEqual('my-role-1');
        expect(data[0].masters[0]).toEqual('my-master-1');
        expect(data[0].masters[1]).toEqual('my-master-2');

        expect(data[1].id).toEqual('my-id-2');
        expect(data[1].server).toEqual('my-server-2');
        expect(data[1].name).toEqual('my-name-2');
        expect(data[1].role).toEqual('my-role-2');
        expect(data[1].masters[0]).toEqual('my-master-2');
        expect(data[1].masters[1]).toEqual('my-master-1');
        done()
    });
});
