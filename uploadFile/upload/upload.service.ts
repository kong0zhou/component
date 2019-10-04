import { Injectable } from '@angular/core';
import { Observable, of } from 'rxjs';
// import { HttpClient } from '@angular/common/http';

@Injectable({
  providedIn: 'root'
})
export class UploadService {

  constructor(
    // private http: HttpClient
  ) { }

  url: string = 'http://106.52.121.11:8084/test'
  upload(files: Set<File>): Map<string, Observable<number>> {
    let allStatus = new Map<string, Observable<number>>()

    files.forEach(file => {
      let formData = new FormData();
      formData.append('file', file, file.name);
      let xhr = new XMLHttpRequest();
      xhr.open("post", this.url, true);
      let sob = new Observable<number>(
        observer => {
          xhr.onload = (evt) => {
            console.log('请求已完成')
            // let s: status = { isFinished: true, perent: 100 }
            observer.complete()
          }
          xhr.onerror = (error) => {
            console.error(error)
            console.error('上传失败')
            observer.error(event)
          }
          xhr.upload.onprogress = (event) => {
            console.log(event.lengthComputable)
            if (event.lengthComputable) {
              let perent = event.loaded / event.total * 100;
              observer.next(perent)
              console.log(perent)
            }
          }
          xhr.upload.onloadstart=()=>{}
          xhr.send(formData)
        }
      )
      allStatus.set(file.name,sob)
    })
    return allStatus
  }
}

class status {
  isFinished: boolean
  perent: number
}