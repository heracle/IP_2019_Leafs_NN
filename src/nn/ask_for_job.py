
import numpy as np
import pandas as pd # pip3 install pandas
import os
import string
import mahotas as mt # pip3 install mahotas

from sklearn.model_selection import train_test_split
from sklearn.preprocessing import StandardScaler
from sklearn import svm
from sklearn import metrics
from sklearn.model_selection import GridSearchCV
from sklearn.decomposition import PCA
import cv2 # pip3 install opencv-python

# ### Testing with mobile captured leaves which are not classified
def bg_sub(filename):
    test_img_path = filename
    main_img = cv2.imread(test_img_path)
    img = cv2.cvtColor(main_img, cv2.COLOR_BGR2RGB)
    resized_image = cv2.resize(img, (1600, 1200))
    size_y,size_x,_ = img.shape
    gs = cv2.cvtColor(resized_image,cv2.COLOR_RGB2GRAY)
    blur = cv2.GaussianBlur(gs, (55,55),0)
    ret_otsu,im_bw_otsu = cv2.threshold(blur,0,255,cv2.THRESH_BINARY_INV+cv2.THRESH_OTSU)
    kernel = np.ones((50,50),np.uint8)
    closing = cv2.morphologyEx(im_bw_otsu, cv2.MORPH_CLOSE, kernel)
    
    contours, hierarchy = cv2.findContours(closing,cv2.RETR_TREE,cv2.CHAIN_APPROX_SIMPLE)
    
    contains = []
    y_ri,x_ri, _ = resized_image.shape
    for cc in contours:
        yn = cv2.pointPolygonTest(cc,(x_ri//2,y_ri//2),False)
        contains.append(yn)

    val = [contains.index(temp) for temp in contains if temp>0]
    index = val[0]
    
    black_img = np.empty([1200,1600,3],dtype=np.uint8)
    black_img.fill(0)
    
    cnt = contours[index]
    mask = cv2.drawContours(black_img, [cnt] , 0, (255,255,255), -1)
    
    maskedImg = cv2.bitwise_and(resized_image, mask)
    white_pix = [255,255,255]
    black_pix = [0,0,0]
    
    final_img = maskedImg
    h,w,channels = final_img.shape
    for x in range(0,w):
        for y in range(0,h):
            channels_xy = final_img[y,x]
            if all(channels_xy == black_pix):
                final_img[y,x] = white_pix
    return final_img

def feature_extract(img):
    names = ['area','perimeter','pysiological_length','pysiological_width','aspect_ratio','rectangularity','circularity',              'mean_r','mean_g','mean_b','stddev_r','stddev_g','stddev_b',              'contrast','correlation','inverse_difference_moments','entropy'
            ]
    df = pd.DataFrame([], columns=names)

    #Preprocessing
    gs = cv2.cvtColor(img,cv2.COLOR_RGB2GRAY)
    blur = cv2.GaussianBlur(gs, (25,25),0)
    ret_otsu,im_bw_otsu = cv2.threshold(blur,0,255,cv2.THRESH_BINARY_INV+cv2.THRESH_OTSU)
    kernel = np.ones((50,50),np.uint8)
    closing = cv2.morphologyEx(im_bw_otsu, cv2.MORPH_CLOSE, kernel)

    #Shape features
    contours, _ = cv2.findContours(closing,cv2.RETR_TREE,cv2.CHAIN_APPROX_SIMPLE)
    cnt = contours[0]
    M = cv2.moments(cnt)
    area = cv2.contourArea(cnt)
    perimeter = cv2.arcLength(cnt,True)
    x,y,w,h = cv2.boundingRect(cnt)
    aspect_ratio = float(w)/h
    rectangularity = w*h/area
    circularity = ((perimeter)**2)/area

    #Color features
    red_channel = img[:,:,0]
    green_channel = img[:,:,1]
    blue_channel = img[:,:,2]
    blue_channel[blue_channel == 255] = 0
    green_channel[green_channel == 255] = 0
    red_channel[red_channel == 255] = 0

    red_mean = np.mean(red_channel)
    green_mean = np.mean(green_channel)
    blue_mean = np.mean(blue_channel)

    red_std = np.std(red_channel)
    green_std = np.std(green_channel)
    blue_std = np.std(blue_channel)

    #Texture features
    textures = mt.features.haralick(gs)
    ht_mean = textures.mean(axis=0)
    contrast = ht_mean[1]
    correlation = ht_mean[2]
    inverse_diff_moments = ht_mean[4]
    entropy = ht_mean[8]

    vector = [area,perimeter,w,h,aspect_ratio,rectangularity,circularity,              
                red_mean,green_mean,blue_mean,red_std,green_std,blue_std,              
                contrast,correlation,inverse_diff_moments,entropy
             ]

    df_temp = pd.DataFrame([vector],columns=names)
    df = df.append(df_temp)
    return df

def get_svm():
    import pickle

    

    pkl_filename = os.path.join(os.environ.get('GOPATH'), 'prediction_data', 'pickle_model.pkl') 
    with open(pkl_filename, 'rb') as file:
        svm_clf = pickle.load(file)
    
    pkl_filename = os.path.join(os.environ.get('GOPATH'), 'prediction_data', 'my_scaler.pkl')
    with open(pkl_filename, 'rb') as file:
        loaded_scaler = pickle.load(file)

    return svm_clf, loaded_scaler

output_filename = None
def finish(output, msg):
    f = open(output_filename, 'w+')
    f.write(output)
    f.close()
    print(msg)
    exit(0)

def main():
    import argparse
    parser = argparse.ArgumentParser(description='Leaf classifier.')
    parser.add_argument('jobid', metavar='JOB ID', type=str, default=None, help='source image job to classify')
    args = parser.parse_args()

    global output_filename
    output_filename = os.path.join(os.environ.get('GOPATH'), 'data_store', args.jobid + '.txt') 
    image_filename = os.path.join(os.environ.get('GOPATH'), 'data_store', args.jobid + '.jpg')

    bg_rem_img = None
    try:
        open(image_filename).close()
        bg_rem_img = bg_sub(image_filename)
        #img = cv2.imread(args.filename)
        #bg_rem_img = cv2.cvtColor(img, cv2.COLOR_BGR2RGB)
    except FileNotFoundError as e:
        finish('-1', e.strerror)
    except IndexError as e:
        finish('-1', str(e) + " bg_sub method fails on contains negative values")
    except Exception as e:
        finish('-1', e)
    
    try:
        svm_clf, scaler = get_svm()
    except Exception as e:
        finish('-1', str(e) + " error retrieving SVM_CLF and SCALER")
    
    try:
        features_of_img = feature_extract(bg_rem_img)
        scaled_features_of_img = scaler.transform(features_of_img)
    except Exception as e:
        finish('-1', str(e) + "error extracting features")    

    try:
        y_pred_mobile = svm_clf.predict(scaled_features_of_img)
    except Exception as e:
        finish('-1', str(e) + " error predicting")
    
    common_names = ['pubescent bamboo','Chinese horse chestnut','Anhui Barberry',
    'Chinese redbud','true indigo','Japanese maple','Nanmu',' castor aralia',
    'Chinese cinnamon','goldenrain tree','Big-fruited Holly','Japanese cheesewood',
    'wintersweet','camphortree','Japan Arrowwood','sweet osmanthus','deodar','ginkgo, maidenhair tree',
    'Crape myrtle, Crepe myrtle','oleander','yew plum pine','Japanese Flowering Cherry','Glossy Privet',
    'Chinese Toon','peach','Ford Woodlotus','trident maple','Beales barberry','southern magnolia',
    'Canadian poplar','Chinese tulip tree','tangerine'
                ]
    # common_names = ['bambus',
    #                 'castan',
    #                 'Dracila',                 
    #                 'Fag',
    #                 'Flori de gladita',
    #                 'Smochin',
    #                 'Nanmu',
    #                 ' castor aralia',                 
    #                 'Chinese cinnamon',
    #                 'goldenrain tree',
    #                 'Big-fruited Holly',
    #                 'Japanese cheesewood',                 
    #                 'wintersweet',
    #                 'camphortree',
    #                 'Japan Arrowwood',
    #                 'sweet osmanthus',
    #                 'deodar',
    #                 'ginkgo, maidenhair tree',                 
    #                 'Crape myrtle, Crepe myrtle',
    #                 'oleander',
    #                 'yew plum pine',
    #                 'Japanese Flowering Cherry',
    #                 'Glossy Privet',                
    #                 'Chinese Toon','peach',
    #                 'Ford Woodlotus',
    #                 'trident maple',
    #                 'Beales barberry',
    #                 'southern magnolia',                
    #                 'Canadian poplar',
    #                 'Chinese tulip tree',
    #                 'tangerine'
    #             ]
    finish(str(y_pred_mobile[0]), 'OK')

if __name__ == '__main__':
    main()
