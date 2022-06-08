import React, { useState, useRef } from 'react';
import { PageContainer, FooterToolbar } from '@ant-design/pro-layout';
import type { ProColumns, ActionType } from '@ant-design/pro-table';
import ProTable from '@ant-design/pro-table';
import { Button, Space, Modal, message } from 'antd';
import { deleteDesignInfo, getDesignList, saveDesignInfo, useDesignInfo } from '@/services/design';
import { history } from 'umi';

const DesignIndex: React.FC = () => {
  const actionRef = useRef<ActionType>();

  const handleUseTemplate = (template: any) => {
    Modal.confirm({
      title: '确定要启用这套设计模板吗？',
      onOk: () => {
        useDesignInfo({package: template.package})
        actionRef.current?.reload();
      }
    })
  }

  const handleManage = (template: any) => {
    history.push('/design/detail?package=' + template.package);
  }

  const handleShowEdit = (template: any) => {
    history.push('/design/editor?package=' + template.package);
  }

  const handleRemove = (packageName: string) => {
    if (packageName == 'default') {
      message.error('默认模板不能删除')
      return;
    }
    Modal.confirm({
      title: '确定要删除这套设计模板吗？',
      onOk: () => {
        deleteDesignInfo({package: packageName}).then(res => {
          message.info(res.msg)
        })
        actionRef.current?.reload();
      }
    })
  }

  const columns: ProColumns<any>[] = [
    {
      title: '名称',
      dataIndex: 'name',
    },
    {
      title: '文件夹',
      dataIndex: 'package',
    },
    {
      title: '类型',
      dataIndex: 'template_type',
      valueEnum: {
        0: '自适应',
        1: '代码适配',
        2: '电脑+手机',
      }
    },
    {
      title: '状态',
      dataIndex: 'status',
      valueEnum: {
        0: {
          text: '未启用',
          status: 'Default',
        },
        1: {
          text: '已启用',
          status: 'Success',
        },
      }
    },
    {
      title: '时间',
      dataIndex: 'created',
    },
    {
      title: '操作',
      key: 'action',
      width: 300,
      render: (text: any, record: any) => (
        <Space size={16}>
          {record.status != 1 && <Button
            type="link"
            onClick={() => {
              handleUseTemplate(record);
            }}
          >
            启用
          </Button>}
          <Button
            type="link"
            onClick={() => {
              handleManage(record);
            }}
          >
            管理
          </Button>
          <Button
            type="link"
            onClick={() => {
              handleShowEdit(record);
            }}
          >
            编辑
          </Button>
          {record.package !== 'default' && <Button
            danger
            type="link"
            onClick={() => {
              handleRemove(record.package);
            }}
          >
            删除
          </Button>}
        </Space>
      ),
    },
  ];

  return (
    <PageContainer>
      <ProTable<any>
        headerTitle="设计模板列表"
        actionRef={actionRef}
        rowKey="package"
        search={false}
        request={(params, sort) => {
          return getDesignList(params);
        }}
        pagination={false}
        columns={columns}
      />
    </PageContainer>
  );
};

export default DesignIndex;
