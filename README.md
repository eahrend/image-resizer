# image-resizer
Image resizing service 

## READ ME FIRST
The helm files are designed for GKE. It will contain ingresses and annotations that won't work elsewhere


## How to install it
1. Connect to your kube cluster
2. Create a static IP address in GCP
3. Run `kubectl create ns <namespace here>`
4. Run `helm install img-resizer ./helm/chart/image-resizer --set ingress.hosts[0]="<your host here>" --namespace=<namespace>`
5. If you want to upgrade to a new image use `helm update img-resizer ./helm/chart/image-resizer --set ingress.hosts[0]="<your host here>" --namespace=<namespace> --set image.name="eahrend/img-resizer:tag"`


## How to use it
1. `/resize/image/<img url>` will return the image with no transformations
2. `/resize/image/<img url>?width=<width>&height=<height>` will modify the image in the url to change height and width


## How to deploy it so it scales
1. The config is "seperated" from the code here, so you could deploy it to any number of kubernetes clusters, no state management needed.
2. Ideally the CI/CD system would listen to dockerhub events so this would automatically deploy


## Steps I took
1. I actually built this before as a proof of concept for a previous company to resize images on the fly rather than resizing images with specific sets on upload. I just modified it to be a lot more agnostic and not use a private s3 bucket. We decided against it since our CDN provider allows us to do this, and will cache those resized images as well.
2. I checked out this code to add a watermark https://www.golangprograms.com/how-to-add-watermark-or-merge-two-image.html
3. I did swap comma separate options to be query parameters so we don't deal with annoying stuff like URL de/encoding by accident.
4. I also wasn't _entirely_ sure on what "quality" meant and how to quantify that. However, if I find out what that means, I can update it.
5. Following the query parameter modification to make it more "standard", the URL format will have changed to be more in line. So the path is `/resizer/image/<URL encoded source of image>?width=1&height=1`
6. Stress testing documentation is located in `./tests/`
   1. Test1/Test2 are user based testing
7. For more accurate stress testing, I'd probably aim to use a proper load testing service, since my macbook started overheating


## Things I'd improve on
1. If we're serving these out of a CDN, odds are there is some sort of system built in that does image optimization
2. Remove this static doc and instead use swagger to make it easier to read the docs
3. Probably look to cache frequently handled images, or set browser side cache header to limit the times a unique browser needs to implement it
4. Actually write tests for this sort of thing, I can test the http server, but I'm trying to limit my time on this project so I don't want to spend too much time on figuring out what I'm expecting out of an image and doing a comparison, though I'm sure it's doable.