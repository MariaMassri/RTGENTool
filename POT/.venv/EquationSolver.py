import numpy as np
from scipy.optimize import nnls
import ctypes
import time
print (ctypes.sizeof(ctypes.c_voidp))

path = "extendedChungLu/"
durationFile = open(path +  'duration3.txt', "a+")
a = open(path + "A.txt", "r")
b = open(path + "B.txt", "r")

pa = a.readlines() #array of equations each equation is a string transform to array of arrays
pb = b.readlines() #array of bs each b is a string

finalA = []
for eq in pa:
    paString = str.split(eq, " ") #array of strings -> array of float64 append to finalA
    pasFloat = []
    for p in paString:
        pasFloat.append(float(p))
    finalA.append(pasFloat)

print("DONE")
print(len((pa)))

finalB = []
for b in pb:
    bf = float(b)
    finalB.append(bf)

print(len(finalB))
print(len(finalA))

#A = np.array(finalA, dtype = object)
#B = np.array(finalB, dtype = object)
A = np.array(finalA, dtype=np.float32)
B = np.array(finalB, dtype=np.float32)
startTime = time.time()
x, rnorm = nnls(np.asmatrix(A),B)
endTime = time.time()
durationFile.write(str(endTime - startTime)+"\n")
durationFile.close()
print (x)
print(len(x))

np.savetxt(path+"pijcm.txt", x ,fmt='%1.7e')