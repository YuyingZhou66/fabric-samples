import numpy as np
import sys
#print(sys.path)
from sklearn.cluster import KMeans
from matplotlib import pyplot
import matplotlib.pyplot as plt
import re


def filtering(s1):
    '''
    This program is for removing all unfair ratings by using k-means clustering
    There are four indicators:
    the feedback on price of product
    the feedback on price of service
    the feedback on quality of service
    the feedback on delivery time

    The result, would separate all input ratings into two classes, unfair rating and fair ratings.
    '''
    s1 = s1.split('],[')
    obRate = []
    for e in s1:
        element = 0
        if e.startswith('['):
            e = e.replace('[', '')
        if e[-1] == ']':
            e = e.replace(']', '')
        row = e.split(',')
        for i in range(0, len(row)):
                row[i] = float(row[i])
        obRate.extend([row])


    ## separate all ratings into two class, and the class with small number is the unfair ratings class
    kmeans = KMeans(n_clusters=2)
    kmeans.fit(obRate)

    centers = kmeans.cluster_centers_.tolist()
    #print("kmeans.cluster_center = ", kmeans.cluster_centers_)

#     marker = ['r*','b*']
#     for i, center in enumerate (centers):
#         plt.plot(center[0], center[1], marker[i],markersize =5)


    labels = kmeans.labels_.tolist()

    # choose the larger cluster be cleaned cluster and keep its label.
    if labels.count(0) > labels.count(1):
        label = 0
    else:
        label = 1

    cleaned_obRateI = []
    # cleaned_obRT = []
    for i, l in enumerate(labels):
        if l == label:
            cleaned_obRateI.append(i)
            #cleaned_obRT.append(obRT[i])

#     mark = ['r.','b.']
#     j=0
#
#     for i in labels:
#         plt.plot([obRate[j][0]], [obRate[j][1]], mark[i],markersize=5)
#         j+=1
#         #print("j=",j)
#     plt.show()

    # the filtered_ratings would be the final result, and will be used in reputation and trust score.
    return cleaned_obRateI