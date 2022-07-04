import numpy as np
import sys 
print(sys.path)

import ot 
import ot.plot
from ot.datasets import make_1D_gauss as gauss
import time

# import matplotlib.pylab as pl 
# import matplotlib.colors as mcolors


path = "extendedChungLu/"
durationFile = open(path +  'duration.txt', "a+")
fs = open(path + "sourceDistribution.txt", "r")
ft = open(path + "targetDistribution.txt", "r")

fs1 = fs.readlines()
ft1 = ft.readlines()
#print(fs1[0])
degSource = fs1[0].split(",")
nbOccSource = fs1[1].split(",")
degTarget = ft1[0].split(",")
nbOccTarget = ft1[1].split(",")

nSource = len(degSource) -1
nTarget = len(degTarget) -1

s = np.zeros((nSource,2))
t = np.zeros((nTarget,2))
nbVertices = 0
nbVerticesTarget = 0
for i in range(0, len(degSource)-1): 
    degSource[i] = float(degSource[i])
for i in range(0, len(nbOccSource)-1): 
    nbOccSource[i] = float(nbOccSource[i]) 
    nbVertices += nbOccSource[i]
for i in range(0, len(degTarget)-1): 
    degTarget[i] = float(degTarget[i]) 
for i in range(0, len(nbOccTarget)-1): 
    nbOccTarget[i] = float(nbOccTarget[i]) 
    nbVerticesTarget += nbOccTarget[i]

degSource.pop()
degTarget.pop()
nbOccSource.pop()
nbOccTarget.pop()

if nbVertices != nbVerticesTarget:
    print("!!!Warning!!! nVertices: " + str(nbVertices) + "!= nbVerticesTarget: "+ str(nbVerticesTarget))
index = len(nbOccSource)-1
for i in range(0, int(nbVertices-nbVerticesTarget)):
    while(nbOccSource[index] - 1 < 0):
        index = index - 1
        nbOccSource[index] = float(nbOccSource[index]) - 1

s[:,0] = degSource
t[:,0] = degTarget
startTime = time.time()

M1 = ot.dist(s, t, metric='euclidean')

print(np.matrix(M1))



M1 /= M1.max()

#a, b = ot.unif(nSource), ot.unif(nTarget) 
a = nbOccSource
print(len(a))

b = nbOccTarget

print(len(b))
a = [x / nbVertices for x in a]
b = [x / nbVertices for x in b]
print(b)
print(a)
#print("nSource")
sumSource = np.sum(a)
#print (a)

sumTarget = np.sum(b)

"""
pl.figure(1, figsize=(7, 3))
pl.clf()
pl.plot(s[:, 0], s[:, 1], '+b', label='Source samples')
pl.plot(t[:, 0], t[:, 1], 'xr', label='Target samples')
pl.axis('equal')
pl.title('Source and Target distributions')


# Cost matrices
pl.figure(2, figsize=(7, 3))

pl.subplot(1, 3, 1)
pl.imshow(M1, interpolation='nearest')
pl.title('Euclidean cost')
"""
G1 = ot.emd(a, b, M1, numItermax=1000000)
#G1 = ot.emd(a, b, M1)
print(len(G1))
distance = ot.emd2(a,b, M1)
f = open(path +  'EMD.txt', "w")
print("EMD Distance"+ str(distance) +"\n")

f.write(str(distance)+"\n")
f.close()
np.savetxt(path + 'OTMATRIX.txt', G1, fmt='%.17f')
endTime = time.time()
durationFile.write(str(endTime - startTime)+"\n")
durationFile.close()

"""
pl.figure(3, figsize=(7, 3))
pl.subplot(1, 3, 1)
s[:, 1] = nbOccSource
t[:, 1] = nbOccTarget
s[:, 1] = [y/1000  for y in s[:, 1]]
t[:, 1] = [y/1000  for y in t[:, 1]]
ot.plot.plot2D_samples_mat(s, t, G1, c=[.5, .5, 1])
pl.plot(s[:, 0], s[:, 1], '+b', label='Source samples')
pl.plot(t[:, 0], t[:, 1], 'xg', label='Target samples')
pl.axis('equal')
pl.legend(loc=0)
pl.title('OT Euclidean')
pl.show()
"""