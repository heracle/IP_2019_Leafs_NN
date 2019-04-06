import sys

jobID = sys.argv[1]
f = open("data_store/" + jobID + ".txt","w+")
f.write("Osal2 Maestre " + jobID + "\n")
