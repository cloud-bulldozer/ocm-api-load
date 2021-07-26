#!/usr/bin/python3

import argparse
import json
import os
import re
import subprocess
import datetime
import pickle
import tarfile
import logging
import plotly.io as pio
import pandas as pd
import sys
from sqlalchemy import create_engine
from docx import Document
from docx.shared import Inches
from googleapiclient.discovery import build
from google_auth_oauthlib.flow import InstalledAppFlow
from google.auth.transport.requests import Request
from googleapiclient.http import MediaFileUpload
from elasticsearch import Elasticsearch, helpers


# Configure Logging
logging.basicConfig(stream=sys.stdout, level=logging.INFO)
logger = logging.getLogger(__name__)

# If modifying these scopes, delete the file token.pickle.
SCOPES = ['https://www.googleapis.com/auth/drive.metadata.readonly',
          'https://www.googleapis.com/auth/drive.file']


def get_gdrive_service(args):
    if args.pickle is None and args.credentials is None:
        logger.error(
            'ERROR: either `--creds` OR `--pickle` options are mandatory')
        exit(1)

    creds = None
    # The file token.pickle stores the user's access and refresh tokens, and is
    # created automatically when the authorization flow completes for the first
    # time.
    if args.pickle is not None and os.path.exists(args.pickle):
        with open(args.pickle, 'rb') as token:
            creds = pickle.load(token)
    # If there are no (valid) credentials available, let the user log in.
    if not creds or not creds.valid:
        if creds and creds.expired and creds.refresh_token:
            creds.refresh(Request())
        else:
            flow = InstalledAppFlow.from_client_secrets_file(
                args.credentials, SCOPES)
            creds = flow.run_local_server(port=0)
        # Save the credentials for the next run
        with open('token.pickle', 'wb') as token:
            pickle.dump(creds, token)
    # return Google Drive API service
    return build('drive', 'v3', credentials=creds)


def generate_graphs(directory):
    graphs = {}
    # Walking through the folder to find all the result files
    # ignoring dirname since we don't need it in this implementation
    for root, _, files in os.walk(directory):
        for filename in files:
            regex = re.compile(r'[\w-]+_([\w-]+).(\w.+)')
            matches = regex.match(filename)

            if 'summaries' not in root and regex.match(filename) and \
                    matches.group(2) == 'json':
                logger.info(f'Generating graph for: {matches.group(1)}')

                # Initializes database for current file in current directory
                # Read by 20000 chunks
                disk_engine = create_engine(
                    'sqlite:///{}/{}.db'.format(root, matches.group(1)))

                j = 0
                index_start = 1
                chunk = 20000
                for df in pd.read_json(root+'/'+filename,
                                       lines=True,
                                       chunksize=chunk):
                    df.index += index_start

                    columns = ['timestamp', 'latency', 'code']

                    for c in df.columns:
                        if c not in columns:
                            df = df.drop(c, axis=1)

                    j += 1
                    logger.info(f'completed {j*chunk} rows')

                    df.to_sql('data', disk_engine, if_exists='append')
                    index_start = df.index[-1] + 1

                df = pd.read_sql_query('SELECT * FROM data', disk_engine)

                data = [{
                    'type': 'scatter',
                    'x': df['timestamp'],
                    'y': df['latency']/1000000,
                    'mode': 'markers',
                    'transforms': [{
                        'type': 'groupby',
                        'groups': df['code'],
                        'color': 'code',
                        'styles': [
                            {'target': '0',
                             'value': {'marker': {'color': 'coral',
                                                  'symbol': 'triangle-down'}}},
                            {'target': '200',
                             'value': {'marker': {'color': 'LightSkyBlue'}}},
                            {'target': '201',
                             'value': {'marker': {'color': 'LightSkyBlue'}}},
                            {'target': '400',
                             'value': {'marker': {'color': 'crimsom',
                                                  'symbol': 'diamond'}}},
                            {'target': '500',
                             'value': {'marker': {'color': 'darkred',
                                                  'symbol': 'diamond-tall'}}}]
                    }]
                }]

                layout = {
                    'title': '<b>Latency per Request: {}</b>'.format(
                        matches.group(1)),
                    'xaxis': {'title': 'Time',
                              'showgrid': 'true',
                              'ticklabelmode': "period"},
                    'yaxis': {'title': 'Milliseconds (log)',
                              'type': 'log'},
                }

                fig_dict = {'data': data, 'layout': layout}

                pio.write_image(fig_dict,
                                root+'/'+matches.group(1)+".png",
                                engine="kaleido",
                                width=1600,
                                height=900,
                                validate=False)
                graphs[matches.group(1)] = root+'/'+matches.group(1)+".png"
                logger.info(f'Graph saved to: {graphs[matches.group(1)]}')
                os.remove('{}/{}.db'.format(root, matches.group(1)))
    return graphs


def show_graphs(directory, filename):
    regex = re.compile(r'(.*/)?[\w-]+_([\w-]+).(\w.+)')
    matches = regex.match(filename)
    if regex.match(filename) and matches.group(3) == 'json':
        # Initializes database for current file in current directory
        # Read by 20000 chunks
        disk_engine = create_engine(
            'sqlite:///{}.db'.format(matches.group(2)))

        j = 0
        index_start = 1
        chunk = 20000
        for df in pd.read_json(os.path.join(directory, filename),
                               lines=True,
                               chunksize=chunk):
            df.index += index_start

            columns = ['timestamp', 'latency', 'code']

            for c in df.columns:
                if c not in columns:
                    df = df.drop(c, axis=1)

            j += 1
            logger.info(f'completed {j*chunk} rows')

            df.to_sql('data', disk_engine, if_exists='append')
            index_start = df.index[-1] + 1

        df = pd.read_sql_query('SELECT * FROM data', disk_engine)

        data = [{
            'type': 'scatter',
            'x': df['timestamp'],
            'y': df['latency']/1000000,
            'mode': 'markers',
            'transforms': [
                {'type': 'groupby',
                 'groups': df['code'],
                 'color': 'code',
                 'styles': [
                    {'target': '0',
                     'value': {'marker': {'color': 'coral',
                                          'symbol': 'triangle-down'}}},
                    {'target': '200',
                     'value': {'marker': {'color': 'LightSkyBlue'}}},
                    {'target': '201',
                     'value': {'marker': {'color': 'LightSkyBlue'}}},
                    {'target': '400',
                     'value': {'marker': {'color': 'crimsom',
                                          'symbol': 'diamond'}}},
                    {'target': '500',
                     'value': {'marker': {'color': 'darkred',
                                          'symbol': 'diamond-tall'}}}]
                 }]
            }]

        layout = {
            'title': '<b>Latency per Request: {}</b>'.format(matches.group(2)),
            'xaxis': {'title': 'Time',
                      'showgrid': 'true',
                      'ticklabelmode': "period"},
            'yaxis': {'title': 'Milliseconds (log)', 'type': 'log'},
        }

        fig_dict = {'data': data, 'layout': layout}

        os.remove('{}.db'.format(matches.group(2)))

        pio.show(fig_dict,
                 engine="kaleido",
                 width=1600,
                 height=900,
                 validate=False)


def generate_summaries(directory):
    try:
        os.stat('{}/summaries'.format(directory))
    except FileNotFoundError:
        os.mkdir('{}/summaries'.format(directory))
    else:
        logger.error('Error with summaries folder.')
        exit(1)

    for root, _, files in os.walk(directory):
        for filename in files:
            regex = re.compile(r'([\w-]+)_([\w-]+).(\w.+)')
            matches = regex.match(filename)

            if 'summaries' not in root and regex.match(filename) and \
                    matches.group(3) == 'json':
                _summary_name = "{}/summaries/{}_{}-summary.json".format(
                               directory,
                               matches.group(1),
                               matches.group(2))
                logger.info(f'Generating summary for: {matches.group(2)}')
                subprocess.run(["vegeta", "report", "--type", "json",
                                "--output",
                                _summary_name,
                                "{}/{}".format(root, filename)])
                logger.info(f'Summary saved to: {_summary_name}')


def read_summaries(directory):
    summaries = {}
    # Walking through the folder to find all the summaries files
    # ignoring dirname since we don't need it in this implementation
    for root, _, files in os.walk(directory):
        for filename in files:
            regex = re.compile(
                r'[\w-]+_([\w-]+)-summary.(\w.+)')
            matches = regex.match(filename)

            if 'summaries' in root and regex.match(filename) and \
                    matches.group(2) == 'json':
                logger.info(f'Reading summary: {filename}')
                df = pd.read_json(root+'/'+filename, lines=True)

                lat = df['latencies'][0]
                summaries[matches.group(1)] = {
                    'requests':     int(df['requests']),
                    'rate':         float(df['rate']),
                    'duration':     int(df['duration']),
                    'min':          int(lat['min']),
                    'mean':         int(lat['mean']),
                    'max':          int(lat['max']),
                    'success':      float(df['success']),
                    'status_codes': df['status_codes'][0],
                    'errors':       df['errors'][0],
                }
    return summaries


def write_docx(directory, summaries, graphs, filename):
    date = datetime.datetime.utcnow()
    document = Document()

    document.add_heading('OCM Performance Test', 0)

    document.add_heading('Test # ', level=1)
    document.add_paragraph('Date: {}'.format(date.strftime("%Y-%m-%d")))

    document.add_heading('Description', level=2)
    document.add_paragraph('The purpose of this test is ...')

    document.add_heading('Notes', level=3)

    document.add_heading('Endpoints', level=2)

    table = document.add_table(rows=1, cols=3)
    hdr_cells = table.rows[0].cells
    hdr_cells[0].text = 'Enpoint'
    hdr_cells[1].text = 'Rate'
    hdr_cells[2].text = 'Notes'
    for r in summaries:
        row_cells = table.add_row().cells
        row_cells[0].text = r
        row_cells[1].text = '{:.2f}/s for {:.2f} minutes'.format(
            summaries[r]['rate'], summaries[r]['duration']/6e10)
        row_cells[2].text = ''

    document.add_heading('Per endpoint data', level=2)
    for r in summaries:
        document.add_heading('{}'.format(r), level=3)
        document.add_picture(graphs[r], width=Inches(16.6), height=Inches(9.4))
        p = document.add_paragraph(
            'Requests\t\tTotal: {}\t\tRate: {:.2f}\n'.format(
                summaries[r]['requests'], summaries[r]['rate']))
        p.add_run(
            'Duration\t\t{:.2f} minutes\n'.format(
                summaries[r]['duration']/6e10))
        p.add_run('Latencies\n')

        document.add_paragraph('Min: {:.4f} ms'.format(
            summaries[r]['min']/1e6), style='List Bullet')
        document.add_paragraph('Mean: {:.4f} ms'.format(
            summaries[r]['mean']/1e6), style='List Bullet')
        document.add_paragraph('Max: {:.4f} ms'.format(
            summaries[r]['max']/1e6), style='List Bullet')

        p2 = document.add_paragraph('Success\t\t{:.2f}%\n'.format(
            summaries[r]['success']*100))
        p2.add_run('Status Codes\t\t\n{}\n'.format(
            summaries[r]['status_codes']))
        p2.add_run('Error Set\t\t\n{}\n'.format(summaries[r]['errors']))
        p2.add_run('Notes').bold = True
        p2.add_run('\n')
        document.add_page_break()
    document.add_heading('Conclusion', level=2)
    document.add_paragraph('Make sure....', style='List Bullet')
    document.add_page_break()
    document.add_heading('Overall Screenshots', level=2)
    if '.docx' not in filename:
        filename = filename+'.docx'
    document.save('{}/{}'.format(directory, filename))


def search(service, query):
    # search for the file
    result = []
    page_token = None
    while True:
        response = service.files().list(q=query,
                                        spaces="drive",
                                        fields="nextPageToken, \
                                            files(id, name, mimeType)",
                                        pageToken=page_token).execute()
        # iterate over filtered files
        for file in response.get("files", []):
            result.append((file["id"], file["name"], file["mimeType"]))
        page_token = response.get('nextPageToken', None)
        if not page_token:
            # no more files
            break
    return result


def create_folder_gdrive(name, service, parent=''):
    """Creates a folder in GDrive if it doesn't exists already.
        - returns the fodler ID of the new or existing folder
    """
    filetype = "application/vnd.google-apps.folder"
    # search for the named folder in GDrive
    search_result = search(
                        service,
                        query="mimeType='{}' and name = '{}' and \
                                trashed = false".format(
                                filetype, name))

    # If it exists return the ID
    if len(search_result) > 0:
        return search_result[0][0]
    else:
        # Create folder
        folder_metadata = {
            "name": name,
            "mimeType": filetype
        }
        # Check if it has parents
        if parent != '':
            folder_metadata['parents'] = [parent]

        file = service.files().create(
                    body=folder_metadata,
                    fields="id").execute()

        return file.get("id")


def upload_files(args):
    """
    Creates a folder and uploads the requests.tar.gz to it
    """
    uuid = ''
    tar = tarfile.open(os.path.join(args.directory, "requests.tar.gz"), "w|gz")
    for root, _, files in os.walk(args.directory):
        for filename in files:
            regex = re.compile(r'([\w-]+)_([\w-]+).(\w.+)')
            matches = regex.match(filename)
            if 'summaries' not in root and regex.match(filename) and \
                    matches.group(3) == 'json':
                uuid = matches.group(1)
                tar.add(os.path.join(args.directory, filename),
                        arcname='requests/{}'.format(filename))
                logger.info('Added file {} to archive'.format(
                            os.path.join(args.directory, filename)))
    tar.close()
    # authenticate account
    service = get_gdrive_service(args)

    parentfoldername = 'ocm-load-tests'
    parentfolderID = create_folder_gdrive(parentfoldername, service)

    UUIDfolderID = create_folder_gdrive(uuid, service, parentfolderID)
    logger.info(f'Folder ID: {UUIDfolderID}')

    # upload requests.tar.g
    # metadata definition
    file_metadata = {
        "name": "requests.tar.gz",
        "parents": [UUIDfolderID]
    }

    # upload
    filename = str(os.path.join(args.directory, "requests.tar.gz"))
    logger.info(f'Uploading file {filename}....')
    media = MediaFileUpload(filename,
                            mimetype='application/gzip',
                            resumable=True)
    file = service.files().create(body=file_metadata,
                                  media_body=media, fields='id').execute()
    logger.info(f'File created, id: {file.get("id")}')


def summarized_requests(path, index_name, test_id, test_name):
    """
    Yields a summarized request document for each line in a given Vegeta
    results file.

    The expected filename format is:
    40f696b8-0258-4a29-99f6-2767bd453548_create-cluster.json
    ^^^                                  ^^^
    Test UUID                            Test Name
    """

    for line in open(path, 'r'):
        req = json.loads(line)
        doc = {
            '_index': index_name,
            'test_name': test_name,
            'uuid': test_id,
            'timestamp': req['timestamp'],
            'code': req['code'],
            'method': req['method'],
            'url': req['url'],
            'latency_ns': req['latency'],
            'bytes_out': req['bytes_out'],
            'bytes_in': req['bytes_in'],
            'has_error': bool(req.get('error')),
            'has_body': bool(req.get('body')),
        }
        yield doc


def push_to_es(args):
    """
    The expected filename format is:
    40f696b8-0258-4a29-99f6-2767bd453548_create-cluster.json
    ^^^                                  ^^^
    Test UUID                            Test Name
    """
    # ElasticSearch Client
    es_host = os.getenv('ES')
    es_index = args.index
    assert es_host, "Did you forget to specify the environment variable `ES`?"
    es = Elasticsearch(es_host, use_ssl=False, verify_certs=False)
    logger.info('Connected to ElasticSearch')
    es.indices.create(index=es_index, ignore=400)  # Ignore IndexAlreadyExists

    for root, _, files in os.walk(args.directory):
        for filename in files:
            regex = re.compile(r'([\w-]+)_([\w-]+).(\w.+)')
            matches = regex.match(filename)
            if 'summaries' not in root and regex.match(filename) and \
                    matches.group(3) == 'json':
                test_id = matches.group(1)
                test_name = matches.group(2)

                logger.info("Indexing result file: %s" % filename)
                helpers.bulk(es, summarized_requests(os.path.join(root,
                                                                  filename),
                                                     es_index,
                                                     test_id,
                                                     test_name))


def main():
    """Automation script to process the results of a test
        - Generate latency spread graphs
        - Generate vegita reports
        - Generate full report
        - Upload results files
    """

    date = datetime.datetime.utcnow()
    parent_parser = argparse.ArgumentParser(add_help=False)
    parent_parser.add_argument('--dir',
                               dest="directory",
                               default='.',
                               required=True,
                               help='directory path were results are stored')
    parent_parser.add_argument('--debug',
                               default=False,
                               required=False,
                               action='store_true',
                               dest="debug",
                               help='debug flag')

    main_parser = argparse.ArgumentParser()

    action_subparsers = main_parser.add_subparsers(title="action",
                                                   dest="action_command")

    graph_parser = action_subparsers.add_parser("graph",
                                                help="generate the graps \
                                                 for the results file",
                                                parents=[parent_parser])

    graph_parser.add_argument('--filename',
                              dest="filename",
                              help='filename of a result to display the graph. \
                                (Overrides generating all graphs.)')

    action_subparsers.add_parser("summary",
                                 help="generates vegeta \
                                  summary for results",
                                 parents=[parent_parser])

    report_parser = action_subparsers.add_parser("report",
                                                 help="generates report",
                                                 parents=[parent_parser])

    report_parser.add_argument('--filename',
                               dest='filename',
                               default='report-{}.docx'.format(
                                                date.strftime("%Y-%m-%d")
                                                ),
                               help='name for the report file.')

    upload_parser = action_subparsers.add_parser("upload",
                                                 help="uploads test results",
                                                 parents=[parent_parser])

    upload_parser.add_argument('--pickle',
                               dest="pickle",
                               help='file with GDrive pickle authorization. \
                                (If not provided, please provide credentials)')

    upload_parser.add_argument('--cred',
                               dest="credentials",
                               help='file with GDrive credentials. \
                                (Ignored if pickle is provided)')

    es_bulk = action_subparsers.add_parser("esbulk",
                                           help="uploads results to ES",
                                           parents=[parent_parser])

    es_bulk.add_argument('--index',
                         dest="index",
                         help='ES index where the documents will be stored.')

    args = main_parser.parse_args()

    if args.action_command == 'graph':
        if args.filename is not None:
            show_graphs(args.directory, args.filename)
        else:
            generate_graphs(args.directory)
    elif args.action_command == 'summary':
        generate_summaries(args.directory)
    elif args.action_command == 'report':
        graphs = generate_graphs(args.directory)
        generate_summaries(args.directory)
        summaries = read_summaries(args.directory)
        write_docx(args.directory, summaries, graphs, args.filename)
    elif args.action_command == 'upload':
        upload_files(args)
    elif args.action_command == 'esbulk':
        push_to_es(args)


if __name__ == "__main__":
    main()
