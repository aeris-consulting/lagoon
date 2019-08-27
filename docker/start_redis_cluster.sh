cluster_config="--appendonly yes \
  --cluster-enabled yes \
  --cluster-require-full-coverage no \
  --cluster-node-timeout 15000 \
  --cluster-migration-barrier 1 \
  --protected-mode no"

for i in {1..6}
do
  port=$((13000+$i))
  data_dir=$(pwd)/cluster-data/data-${i}
  mkdir -p ${data_dir}
  nohup redis-server --port ${port} --dir ${data_dir} --cluster-config-file ${data_dir}/nodes.conf $cluster_config >> ${data_dir}/redis-server.log &
done

# If you need to create a cluster and load a unique backup, follow these instructions:

# 1. Create the cluster:
# > redis-cli -p 13001 --cluster create --cluster-replicas 1 $(for i in {1..6}; do echo -n "127.0.0.1:$((13000+$i)) " ; done)

# 2. Reshard all to first master
# > redis-cli -p 13001 --cluster reshard 127.0.0.1:13001
# How many slots do you want to move (from 1 to 16384)? 16384
# What is the receiving node ID? 0b0675c3647e3c7e534ca948d472b71c343bae36
# Please enter all the source node IDs.
#   Type 'all' to use all the nodes as source nodes for the hash slots.
#   Type 'done' once you entered all the source nodes IDs.
# Source node #1: all


# 3. Stop the nodes and copy the data (DUMP or APPEND file) on the server and start it again.
# Wait that the data is fully loaded before you go to next step, this might take minutes.
# > ps -aef | grep redis-server | grep -v grep | awk '{print $2}' | xargs kill -15

# 4. Check that data are loaded
# > redis-cli -p 13001 --scan --pattern * | head -n 10

# 5. Rebalance slots between nodes 1,2,3
# To rebalance the shards on the second node, select the third of the maximal number of slots (5461) and assign them to the second master.
# To rebalance the shards on the third node, select the third of the maximal number of slots (5461) and assign them to the third master but only select the first master as source!
# > redis-cli -p 13001 --cluster reshard 127.0.0.1:13001

# 6. Check that slots are balanced.
# > redis-cli -p 13001 cluster info
# > redis-cli -p 13001 cluster nodes