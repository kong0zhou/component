import { Component, OnInit, ViewChild, ElementRef } from '@angular/core';
import { MatDialogRef } from '@angular/material';
import { UploadService } from '../../upload.service'
import { Observable, of } from 'rxjs';

@Component({
  selector: 'app-dialog',
  templateUrl: './dialog.component.html',
  styleUrls: ['./dialog.component.css']
})
export class DialogComponent implements OnInit {
  @ViewChild('file', { static: false }) file: ElementRef;
  constructor(
    public dialogRef: MatDialogRef<DialogComponent>,
    public service: UploadService
  ) { }

  ngOnInit() {
  }

  addFiles() {
    let addfiles = <HTMLInputElement>(this.file.nativeElement)
    addfiles.click()
    console.log(123)

  }

  cancel() {
    this.dialogRef.close();
  }

  // 集合不能有相同的文件
  files: Set<File> = new Set<File>();

  onFilesAdded() {
    console.log(456)
    let files: { [key: string]: File } = this.file.nativeElement.files
    for (let key in files) {
      if (!isNaN(parseInt(key))) {
        this.files.add(files[key])
        // files[key].
      }
    }
  }
  // 是否正在上传
  uploading: boolean = false
  // 是否上传完成
  uploadOver: boolean = false

  perents: Map<string,number> =new Map<string,number>()
  upload() {
    this.uploading = true;
    console.log(789)
    let obs= this.service.upload(this.files);
    let overNum:number = 0
    obs.forEach(
      (value, key) => {
        value.subscribe(
          (data)=>{
            this.perents.set(key,data)
          },
          (error)=>{
            console.error(error)
          },
          ()=>{
            overNum++
            if (overNum===obs.size) {
              this.uploadOver=true;
              console.log('全部数据完成传送')
              this.uploading=false;
            }
          }
        )
      }
    )
  }
}