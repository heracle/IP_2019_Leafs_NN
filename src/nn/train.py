
import numpy as np
import pandas as pd # pip3 install pandas
import os
import string
import matplotlib.pyplot as plt
import mahotas as mt # pip3 install mahotas

from sklearn.model_selection import train_test_split
from sklearn.preprocessing import StandardScaler
from sklearn import svm
from sklearn import metrics
from sklearn.model_selection import GridSearchCV
from sklearn.decomposition import PCA
import cv2 # pip3 install opencv-python

# ### Reading the dataset
dataset = pd.read_csv("train_data/Flavia_features_downloaded.csv")

# print(dataset.head(5))
# print(type(dataset))

# no need to touch images though
# ds_path = '../Flavia leaves dataset'
# img_files = os.listdir(ds_path)

img_files = []
for i in range(1001, 3622):
    img_files.append("train_data/" + str(i) + ".jpg")

# ### Creating target labels
# 
# Breakpoints are used alongside the image file to create a vector of target labels. The breakpoints are specified in Flavia leaves dataset website.

breakpoints = [1001,1059,1060,1122,1552,1616,1123,1194,1195,1267,1268,1323,1324,1385,1386,1437,1497,1551,1438,1496,2001,2050,2051,2113,2114,2165,2166,2230,2231,2290,2291,2346,2347,2423,2424,2485,2486,2546,2547,2612,2616,2675,3001,3055,3056,3110,3111,3175,3176,3229,3230,3281,3282,3334,3335,3389,3390,3446,3447,3510,3511,3563,3566,3621]

target_list = []
for file in img_files:
    target_num = int(file.split(".")[0].split("/")[-1])
    flag = 0
    i = 0 
    for i in range(0,len(breakpoints),2):
        if((target_num >= breakpoints[i]) and (target_num <= breakpoints[i+1])):
            flag = 1
            break
    if(flag==1):
        target = int((i/2))
        target_list.append(target)

y = np.array(target_list)
# print(y)

X = dataset.iloc[:,1:]
# print(X.head(5))
# print(y[0:5])

# ### Train test split

X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.3, random_state = 142)
# print(X_train.head(5))
# print(y_train[0:5])

# ### Feature Scaling

sc_X = StandardScaler()
X_train = sc_X.fit_transform(X_train)
X_test = sc_X.transform(X_test)

# print(X_train[0:2])
# print(y_train[0:2])

# ### Applying SVM classifier model

print('Starting basic training...')
clf = svm.SVC()
clf.fit(X_train,y_train)
print('Basic training finished.')

y_pred = clf.predict(X_test)

metrics.accuracy_score(y_test, y_pred)

# Bad 80% classifier results
# print(metrics.classification_report(y_test, y_pred))

# ### Performing parameter tuning of the model

parameters = [{'kernel': ['rbf'],
            'gamma': [1e-4, 1e-3, 0.01, 0.1, 0.2, 0.5],
            'C': [1, 10, 100, 1000]},
            {'kernel': ['linear'], 'C': [1, 10, 100, 1000]}
            ]

svm_clf = GridSearchCV(svm.SVC(decision_function_shape='ovr'), parameters, cv=5)
print('Starting advanced training...')
svm_clf.fit(X_train, y_train)
print('Advanced training finished.')

# print(svm_clf.best_params_)

means = svm_clf.cv_results_['mean_test_score']
stds = svm_clf.cv_results_['std_test_score']
for mean, std, params in zip(means, stds, svm_clf.cv_results_['params']):
    # print("%0.3f (+/-%0.03f) for %r" % (mean, std * 2, params))
    pass

y_pred_svm = svm_clf.predict(X_test)

#print(y_pred_svm)
#plt.hist(y_pred_svm, bins='auto')
#plt.show()
#exit(0)

metrics.accuracy_score(y_test, y_pred_svm)

# Good 90% classifier results
# print(metrics.classification_report(y_test, y_pred_svm))

# ### Dimensionality Reduction using PCA

pca = PCA()
pca.fit(X)

var = pca.explained_variance_ratio_
# print(var)

var1=np.cumsum(np.round(pca.explained_variance_ratio_, decimals=4)*100)
# plt.plot(var1)
# plt.show()

import pickle
import os

if not os.path.exists('prediction_data'):
    os.makedirs('prediction_data')

pkl_filename = "prediction_data/pickle_model.pkl" 
with open(pkl_filename, 'wb') as file:  
    pickle.dump(svm_clf, file)

pkl_filename = "prediction_data/my_scaler.pkl"
with open(pkl_filename, 'wb') as file:  
    pickle.dump(sc_X, file)
