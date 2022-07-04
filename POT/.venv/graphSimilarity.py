import networkx as nx

G1 = nx.cycle_graph(6)
G2 = nx.wheel_graph(7)
for v in nx.optimize_graph_edit_distance(G1, G2):
    minv = v
print(minv)