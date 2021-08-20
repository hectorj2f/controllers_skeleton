package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)


// ImageScanSpec defines the desired state of ImageScan
type ImageScanSpec struct {

}

// ImageScanStatus defines the observed state of ImageScan
type ImageScanStatus struct {


//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="ScannedImage",type="string",JSONPath=".status.scannedImage"

// ImageScan is the Schema for the imagescans API
type ImageScan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

//+kubebuilder:object:root=true

// ImageScanList contains a list of ImageScan
type ImageScanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ImageScan `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ImageScan{}, &ImageScanList{})
}
