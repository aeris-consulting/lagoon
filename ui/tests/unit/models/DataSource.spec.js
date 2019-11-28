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
        })
        done()
    })
});

test.only('parsing cluster nodes information', (done) => {
    axios.post.mockImplementation(() => {
        return Promise.resolve({
            status: 200,
            data: {
                data: `a9845ec8e989f835e45893011d41bc7f451db740 127.0.0.1:7004 slave 59cd91942c9f8a239cb0ad93aec2c38057d20596 0 1572785824611 3 connected
59cd91942c9f8a239cb0ad93aec2c38057d20596 127.0.0.1:7001 master - 0 1572785825618 1 connected 5462-10923
b649a0ab60323e9463208a43926800aee39799b3 127.0.0.1:7002 master - 0 1572785825114 5 connected 10924-16383
1c5684ec22485074dd63cb49a544bd40cd36ac8f 127.0.0.1:7003 slave 6f57b139637c2047cd257974f426585a41211123 0 1572785825618 4 connected
f7d4f909952f1cc1574e8de4513d9f674f20ae31 127.0.0.1:7005 slave b649a0ab60323e9463208a43926800aee39799b3 0 1572785826123 5 connected
6f57b139637c2047cd257974f426585a41211123 127.0.0.1:7000@1700 myself,master - 0 0 2 connected 0-5461`
            }
        })
    })
    dataSource.getClusterNodes().then((data) => {
        expect(data[0].id).toEqual('a9845ec8e989f835e45893011d41bc7f451db740')
        expect(data[0].ip).toEqual('127.0.0.1:7004')
        expect(data[0].role).toEqual('slave')
        expect(data[5].ip).toEqual('127.0.0.1:7000')
        expect(data[5].role).toEqual('master')
        done()
    });
})
