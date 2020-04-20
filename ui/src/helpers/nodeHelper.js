export default class NodeHelper {
    /**
     * Remove the node from its parent if it has neither content nor children.
     */
    static removeFromParentIfEmpty(node) {
        if (!node.hasChildren && !node.hasContent && node.parent) {
            let deletedNodeIndex = node.parent.children.findIndex(c => c.fullPath === node.fullPath);
            node.parent.children.splice(deletedNodeIndex, 1)
        }
    }
}