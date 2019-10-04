import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { UploadComponent } from './upload/upload.component';
import { MatButtonModule, MatDialogModule, MatListModule, MatProgressBarModule,} from '@angular/material';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { DialogComponent } from './upload/dialog/dialog.component';
import { HttpClientModule } from '@angular/common/http';


@NgModule({
  declarations: [UploadComponent, DialogComponent],
  imports: [
    CommonModule,
    MatButtonModule,
    MatDialogModule,
    MatListModule,
    MatProgressBarModule,
    BrowserAnimationsModule,
    HttpClientModule,
  ],
  exports: [
    UploadComponent,
    // DialogComponent
  ],
  entryComponents: [DialogComponent],
})
export class UploadModule { }
