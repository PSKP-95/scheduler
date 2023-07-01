import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Schedule, SchedulesResponse } from '../models/schedule.model';
import { Observable } from 'rxjs';
import {map} from 'rxjs/operators';

@Injectable({
  providedIn: 'root'
})
export class SchedulesService {
  private scheduleUrl = "/api/schedules"
  schedules: Schedule[];

  constructor(private http: HttpClient) { }

  getSchedules(size: number, page: number): Observable<SchedulesResponse> {
    return this.http.get<any>(`${this.scheduleUrl}?page=${page}&size=${size}`)
      .pipe(
        map(value => this.schedules = value)
      );
  }
}
