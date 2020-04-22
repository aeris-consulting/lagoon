import NodeHelper from '../../../src/helpers/nodeHelper'

describe('NodeHelper', () => {
    it('should remove the empty node from the parent', () => {
        const parentNode = {
            hasChildren: true,
            name: 'parent',
            fullPath: 'parent',
            level: 0
        };
        const nodeToRemove = {
            parent: parentNode,
            hasContent: false,
            hasChildren: false,
            name: 'childToRemove',
            fullPath: 'childToRemove',
            level: 1
        };
        parentNode.children = [{fullPath: 'child-1'}, nodeToRemove, {fullPath: 'child-2'}];
        NodeHelper.removeFromParentIfEmpty(nodeToRemove);
        expect(parentNode.children.find(c => c.fullPath === 'childToRemove')).toBeUndefined();
    })

    it('should not throw an error when there is no parent', () => {
        const nodeToRemove = {
            hasContent: false,
            hasChildren: false,
            name: 'childToRemove',
            fullPath: 'childToRemove',
            level: 1
        };
        NodeHelper.removeFromParentIfEmpty(nodeToRemove);
    })

    it('should not remove anything when the node still has a content', () => {
        const parentNode = {
            hasChildren: true,
            name: 'parent',
            fullPath: 'parent',
            level: 0
        };
        const nodeToRemove = {
            parent: parentNode,
            hasContent: true,
            hasChildren: false,
            name: 'childToRemove',
            fullPath: 'childToRemove',
            level: 1
        };
        parentNode.children = [{fullPath: 'child-1'}, nodeToRemove, {fullPath: 'child-2'}];
        NodeHelper.removeFromParentIfEmpty(nodeToRemove);
        expect(parentNode.children.findIndex(c => c.fullPath === 'childToRemove')).toBe(1);
    })

    it('should not remove anything when the node still has children', () => {
        const parentNode = {
            hasChildren: true,
            name: 'parent',
            fullPath: 'parent',
            level: 0
        };
        const nodeToRemove = {
            parent: parentNode,
            hasContent: false,
            hasChildren: true,
            name: 'childToRemove',
            fullPath: 'childToRemove',
            level: 1
        };
        parentNode.children = [{fullPath: 'child-1'}, nodeToRemove, {fullPath: 'child-2'}];
        NodeHelper.removeFromParentIfEmpty(nodeToRemove);
        expect(parentNode.children.findIndex(c => c.fullPath === 'childToRemove')).toBe(1);
    })

})
