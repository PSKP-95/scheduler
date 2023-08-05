import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Schedule, SchedulesResponse } from '../models/schedule.model';
import { Observable } from 'rxjs';
import {map} from 'rxjs/operators';

@Injectable({
  providedIn: 'root'
})
export class SchedulesService {
  private scheduleUrl = "/api/schedules";
  private triggerScheduleUrl = "/api/schedule/";
  private hooksUrl = "/api/hooks";
  schedules: Schedule[];

  constructor(private http: HttpClient) { }

  getSchedules(size: number, page: number): Observable<SchedulesResponse> {
    return this.http.get<any>(`${this.scheduleUrl}?page=${page}&size=${size}`)
      .pipe(
        map(value => this.schedules = value)
      );
  }

  triggerSchedule(id: string) {
    return this.http.get<any>(`${this.triggerScheduleUrl}/${id}/trigger`)
      .pipe(
        map(value => console.log(value))
      )
  }

  getHooks(): Observable<string[]> {
    return this.http.get<any>(this.hooksUrl);
  }
}
